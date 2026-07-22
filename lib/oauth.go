package lib

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// Gmail defaults. Other providers (e.g. Outlook) can be used by overriding
// oauth_auth_url / oauth_token_url / oauth_scope in the config.
const (
	defaultOAuthAuthURL  = "https://accounts.google.com/o/oauth2/auth"
	defaultOAuthTokenURL = "https://oauth2.googleapis.com/token"
	defaultOAuthScope    = "https://mail.google.com/"
)

// oobRedirectURL is the redirect URI used by the headless setup flow. The user
// copies the code out of the browser by hand, so nothing needs to listen on the
// machine running the setup. It must be registered in the provider's console.
const oobRedirectURL = "urn:ietf:wg:oauth:2.0:oob"

// Xoauth2Client implements sasl.Client for the XOAUTH2 mechanism, which Gmail
// and Outlook use for IMAP. go-sasl only ships OAUTHBEARER (RFC 7628), which
// Gmail does not accept, so the mechanism is implemented here.
//
// See https://developers.google.com/gmail/imap/xoauth2-protocol
type Xoauth2Client struct {
	Username string
	Token    string
}

// Start returns the mechanism name and the initial client response.
func (a *Xoauth2Client) Start() (string, []byte, error) {
	ir := []byte("user=" + a.Username + "\x01auth=Bearer " + a.Token + "\x01\x01")
	return "XOAUTH2", ir, nil
}

// Next handles the server challenge. On failure the server sends a base64 JSON
// error blob and expects an empty response before returning the tagged NO, so
// return the challenge as an error rather than an empty response.
func (a *Xoauth2Client) Next(challenge []byte) ([]byte, error) {
	return nil, fmt.Errorf("XOAUTH2 authentication failed: %s", strings.TrimSpace(string(challenge)))
}

// UseOAuth2 returns whether the config is set to authenticate via OAuth2.
func (c YamlConfig) UseOAuth2() bool {
	return strings.EqualFold(strings.TrimSpace(c.Auth), authOAuth2)
}

// oauthConfig builds the oauth2.Config from the yaml config, applying the Gmail
// defaults for any endpoint the user has not overridden.
func (c YamlConfig) oauthConfig(redirectURL string) *oauth2.Config {
	authURL := c.OAuthAuthURL
	if authURL == "" {
		authURL = defaultOAuthAuthURL
	}
	tokenURL := c.OAuthTokenURL
	if tokenURL == "" {
		tokenURL = defaultOAuthTokenURL
	}
	scope := c.OAuthScope
	if scope == "" {
		scope = defaultOAuthScope
	}

	return &oauth2.Config{
		ClientID:     c.OAuthClientID,
		ClientSecret: c.OAuthClientSecret,
		Scopes:       []string{scope},
		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}
}

// LoadOAuthToken reads the stored token from oauth_token_file.
func LoadOAuthToken(file string) (*oauth2.Token, error) {
	file = path.Clean(file)
	// #nosec G304 -- path comes from the user's own config file
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	tok := &oauth2.Token{}
	if err := json.Unmarshal(data, tok); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", file, err)
	}
	if tok.RefreshToken == "" {
		return nil, fmt.Errorf("%s contains no refresh_token", file)
	}

	return tok, nil
}

// SaveOAuthToken writes the token to oauth_token_file with 0600 permissions.
func SaveOAuthToken(file string, tok *oauth2.Token) error {
	file = path.Clean(file)
	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(file, append(data, '\n'), 0o600)
}

// OAuthAccessToken returns a valid access token for the IMAP login. It only
// reads oauth_token_file and refreshes against the token endpoint, so it is
// safe on headless hosts — no browser is ever opened. A rotated refresh token
// is written back to disk so the next run keeps working.
func OAuthAccessToken() (string, error) {
	stored, err := LoadOAuthToken(Config.OAuthTokenFile)
	if err != nil {
		return "", fmt.Errorf("%w (run --oauth-setup to authorize)", err)
	}

	conf := Config.oauthConfig("")
	fresh, err := conf.TokenSource(context.Background(), stored).Token()
	if err != nil {
		return "", fmt.Errorf("refreshing OAuth2 token: %w", err)
	}

	// Providers may hand back a new refresh token; persist it or the stored one
	// eventually goes stale and login starts failing.
	if fresh.RefreshToken != "" && fresh.RefreshToken != stored.RefreshToken {
		if err := SaveOAuthToken(Config.OAuthTokenFile, fresh); err != nil {
			Log.Warningf("could not save refreshed OAuth2 token: %s", err)
		}
	}

	return fresh.AccessToken, nil
}

// OAuthSetup runs the one-off authorisation flow and writes the resulting
// refresh token to oauth_token_file. When headless is true the URL is printed
// for the user to open elsewhere and the code is pasted back in; otherwise a
// loopback listener catches the redirect and the browser is opened locally.
func OAuthSetup(headless bool) error {
	if headless {
		return oauthSetupHeadless()
	}

	return oauthSetupInteractive()
}

// oauthSetupHeadless prints the authorization URL and reads the resulting code
// from stdin. Nothing listens locally, so the user can complete the browser
// half of the flow on any other machine.
func oauthSetupHeadless() error {
	conf := Config.oauthConfig(oobRedirectURL)
	authURL := conf.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	fmt.Printf("\nOpen this URL in a browser on any machine and grant access:\n\n%s\n\n", authURL)
	fmt.Printf("Then paste the authorization code (or the full redirect URL) here: ")

	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading authorization code: %w", err)
	}

	code := parseAuthCode(strings.TrimSpace(line))
	if code == "" {
		return errors.New("no authorization code entered")
	}

	return exchangeAndSave(conf, code)
}

// oauthSetupInteractive starts a loopback listener, opens the browser on this
// machine and waits for the provider to redirect back with the code.
func oauthSetupInteractive() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("starting local redirect listener: %w", err)
	}
	defer listener.Close() // nolint:errcheck

	redirectURL := fmt.Sprintf("http://127.0.0.1:%d", listener.Addr().(*net.TCPAddr).Port)
	conf := Config.oauthConfig(redirectURL)
	authURL := conf.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	type result struct {
		code string
		err  error
	}
	results := make(chan result, 1)

	server := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			if errMsg := query.Get("error"); errMsg != "" {
				http.Error(w, "Authorization failed: "+errMsg, http.StatusBadRequest)
				results <- result{err: fmt.Errorf("authorization denied: %s", errMsg)}
				return
			}

			code := query.Get("code")
			if code == "" {
				// Ignore unrelated requests such as /favicon.ico.
				http.NotFound(w, r)
				return
			}

			fmt.Fprintln(w, "Authorization complete. You can close this window and return to the terminal.")
			results <- result{code: code}
		}),
	}
	//nolint:errcheck
	go server.Serve(listener)
	defer server.Close() // nolint:errcheck

	fmt.Printf("\nOpening a browser to authorize access. If it does not open, visit:\n\n%s\n\n", authURL)
	if err := openBrowser(authURL); err != nil {
		Log.DebugF("Could not open a browser automatically: %s", err)
	}

	select {
	case res := <-results:
		if res.err != nil {
			return res.err
		}
		return exchangeAndSave(conf, res.code)
	case <-time.After(5 * time.Minute):
		return errors.New("timed out waiting for authorization, try -oauth-headless instead")
	}
}

// exchangeAndSave swaps the authorization code for a token and stores it.
func exchangeAndSave(conf *oauth2.Config, code string) error {
	tok, err := conf.Exchange(context.Background(), code)
	if err != nil {
		return fmt.Errorf("exchanging authorization code: %w", err)
	}

	if tok.RefreshToken == "" {
		return errors.New("provider returned no refresh_token; revoke the app's access and run setup again")
	}

	if err := SaveOAuthToken(Config.OAuthTokenFile, tok); err != nil {
		return fmt.Errorf("saving token: %w", err)
	}

	Log.NoticeF("Saved OAuth2 refresh token to %s", Config.OAuthTokenFile)

	return nil
}

// parseAuthCode accepts either a bare authorization code or the whole redirect
// URL the browser landed on, and returns the code.
func parseAuthCode(input string) string {
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return input
	}

	u, err := url.Parse(input)
	if err != nil {
		return input
	}

	return u.Query().Get("code")
}

// openBrowser tries to open url in the user's default browser.
func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		// #nosec G204 -- url is built from the config's own OAuth endpoint
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		// #nosec G204
		return exec.Command("open", url).Start()
	default:
		// #nosec G204
		return exec.Command("xdg-open", url).Start()
	}
}

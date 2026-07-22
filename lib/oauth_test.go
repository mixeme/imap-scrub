package lib

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestXoauth2ClientStart(t *testing.T) {
	c := &Xoauth2Client{Username: "user@example.com", Token: "ya29.token"}

	mech, ir, err := c.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if mech != "XOAUTH2" {
		t.Fatalf("mech = %q, want XOAUTH2", mech)
	}

	// Format is fixed by the Gmail XOAUTH2 spec; a stray byte breaks login.
	want := "user=user@example.com\x01auth=Bearer ya29.token\x01\x01"
	if string(ir) != want {
		t.Fatalf("ir = %q, want %q", ir, want)
	}

	// Sanity check that it survives the base64 the IMAP client applies.
	if _, err := base64.StdEncoding.DecodeString(base64.StdEncoding.EncodeToString(ir)); err != nil {
		t.Fatalf("initial response is not base64-safe: %v", err)
	}
}

func TestXoauth2ClientNextIsAlwaysAnError(t *testing.T) {
	c := &Xoauth2Client{Username: "user@example.com", Token: "bad"}

	// The server only sends a challenge when auth failed.
	if _, err := c.Next([]byte(`{"status":"401"}`)); err == nil {
		t.Fatal("Next() expected an error on server challenge")
	}
}

func TestLoadOAuthToken(t *testing.T) {
	file := filepath.Join(t.TempDir(), "token.json")
	if err := os.WriteFile(file, []byte(`{"access_token":"at","refresh_token":"rt"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	tok, err := LoadOAuthToken(file)
	if err != nil {
		t.Fatalf("LoadOAuthToken() error = %v", err)
	}
	if tok.RefreshToken != "rt" {
		t.Fatalf("RefreshToken = %q, want rt", tok.RefreshToken)
	}
}

func TestLoadOAuthTokenRejectsMissingRefreshToken(t *testing.T) {
	file := filepath.Join(t.TempDir(), "token.json")
	if err := os.WriteFile(file, []byte(`{"access_token":"at"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadOAuthToken(file); err == nil {
		t.Fatal("LoadOAuthToken() expected an error for a token with no refresh_token")
	}
}

func TestLoadOAuthTokenMissingFile(t *testing.T) {
	if _, err := LoadOAuthToken(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Fatal("LoadOAuthToken() expected an error for a missing file")
	}
}

func TestSaveOAuthTokenRoundTrip(t *testing.T) {
	file := filepath.Join(t.TempDir(), "token.json")
	want := &oauth2.Token{AccessToken: "at", RefreshToken: "rt", Expiry: time.Now().Add(time.Hour)}

	if err := SaveOAuthToken(file, want); err != nil {
		t.Fatalf("SaveOAuthToken() error = %v", err)
	}

	got, err := LoadOAuthToken(file)
	if err != nil {
		t.Fatalf("LoadOAuthToken() error = %v", err)
	}
	if got.AccessToken != want.AccessToken || got.RefreshToken != want.RefreshToken {
		t.Fatalf("round trip = %+v, want %+v", got, want)
	}
}

func TestParseAuthCode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"bare code", "4/0Ab_code", "4/0Ab_code"},
		{"redirect url", "http://127.0.0.1:5000/?state=state&code=4/0Ab_code&scope=mail", "4/0Ab_code"},
		{"https redirect url", "https://example.com/cb?code=abc", "abc"},
		{"url without a code", "http://127.0.0.1:5000/?error=access_denied", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseAuthCode(tc.input); got != tc.want {
				t.Fatalf("parseAuthCode(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestOAuthConfigDefaultsToGmail(t *testing.T) {
	cfg := YamlConfig{OAuthClientID: "id", OAuthClientSecret: "secret"}

	conf := cfg.oauthConfig("http://127.0.0.1:1234")
	if conf.Endpoint.AuthURL != defaultOAuthAuthURL || conf.Endpoint.TokenURL != defaultOAuthTokenURL {
		t.Fatalf("endpoint = %+v, want Gmail defaults", conf.Endpoint)
	}
	if len(conf.Scopes) != 1 || conf.Scopes[0] != defaultOAuthScope {
		t.Fatalf("Scopes = %v, want [%s]", conf.Scopes, defaultOAuthScope)
	}
	if conf.RedirectURL != "http://127.0.0.1:1234" {
		t.Fatalf("RedirectURL = %q", conf.RedirectURL)
	}
}

func TestOAuthConfigHonoursOverrides(t *testing.T) {
	cfg := YamlConfig{
		OAuthAuthURL:  "https://login.example.com/authorize",
		OAuthTokenURL: "https://login.example.com/token",
		OAuthScope:    "https://outlook.office.com/IMAP.AccessAsUser.All",
	}

	conf := cfg.oauthConfig("")
	if conf.Endpoint.AuthURL != cfg.OAuthAuthURL || conf.Endpoint.TokenURL != cfg.OAuthTokenURL {
		t.Fatalf("endpoint = %+v, want the configured overrides", conf.Endpoint)
	}
	if conf.Scopes[0] != cfg.OAuthScope {
		t.Fatalf("Scopes = %v, want [%s]", conf.Scopes, cfg.OAuthScope)
	}
}

// OAuthAccessToken must refresh against the token endpoint and persist a
// rotated refresh token, otherwise the stored one goes stale.
func TestOAuthAccessTokenRefreshesAndPersistsRotatedToken(t *testing.T) {
	var gotRefreshToken string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm() error = %v", err)
		}
		gotRefreshToken = r.Form.Get("refresh_token")
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck
		w.Write([]byte(`{"access_token":"new-access","refresh_token":"rotated","token_type":"Bearer","expires_in":3600}`))
	}))
	defer server.Close()

	file := filepath.Join(t.TempDir(), "token.json")
	if err := os.WriteFile(file, []byte(`{"access_token":"old","refresh_token":"original","expiry":"2000-01-01T00:00:00Z"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	defer withConfig(YamlConfig{
		Auth:              authOAuth2,
		User:              "user@example.com",
		OAuthClientID:     "id",
		OAuthClientSecret: "secret",
		OAuthTokenFile:    file,
		OAuthTokenURL:     server.URL,
	})()

	token, err := OAuthAccessToken()
	if err != nil {
		t.Fatalf("OAuthAccessToken() error = %v", err)
	}
	if token != "new-access" {
		t.Fatalf("token = %q, want new-access", token)
	}
	if gotRefreshToken != "original" {
		t.Fatalf("server received refresh_token %q, want original", gotRefreshToken)
	}

	stored, err := LoadOAuthToken(file)
	if err != nil {
		t.Fatalf("LoadOAuthToken() error = %v", err)
	}
	if stored.RefreshToken != "rotated" {
		t.Fatalf("stored RefreshToken = %q, want rotated", stored.RefreshToken)
	}
}

func TestOAuthAccessTokenReportsRefreshFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
	}))
	defer server.Close()

	file := filepath.Join(t.TempDir(), "token.json")
	if err := os.WriteFile(file, []byte(`{"refresh_token":"revoked","expiry":"2000-01-01T00:00:00Z"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	defer withConfig(YamlConfig{
		Auth:              authOAuth2,
		OAuthClientID:     "id",
		OAuthClientSecret: "secret",
		OAuthTokenFile:    file,
		OAuthTokenURL:     server.URL,
	})()

	_, err := OAuthAccessToken()
	if err == nil {
		t.Fatal("OAuthAccessToken() expected an error when the refresh is rejected")
	}
	if !strings.Contains(err.Error(), "refreshing OAuth2 token") {
		t.Fatalf("error = %v, want it to mention the refresh", err)
	}
}

// withConfig swaps the package-global Config and returns a restore func.
func withConfig(cfg YamlConfig) func() {
	previous := Config
	Config = cfg

	return func() { Config = previous }
}

package lib

import (
	"errors"
	"net"
	"strings"
	"testing"

	"github.com/emersion/go-imap/backend/memory"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap/server"
	"github.com/emersion/go-sasl"
)

// xoauth2Server is a minimal server-side XOAUTH2 mechanism used to check that
// Authenticate() drives a real IMAP AUTHENTICATE exchange correctly. The unit
// tests assert the byte format in isolation; this asserts the server actually
// accepts it over the wire, which is where a formatting bug would really bite.
type xoauth2Server struct {
	wantUser  string
	wantToken string
	gotRaw    string
}

func (s *xoauth2Server) Next(response []byte) ([]byte, bool, error) {
	s.gotRaw = string(response)

	user, token, err := parseXoauth2(s.gotRaw)
	if err != nil {
		return nil, false, err
	}
	if user != s.wantUser || token != s.wantToken {
		return nil, false, errors.New("bad credentials")
	}

	return nil, true, nil
}

// parseXoauth2 decodes "user=<user>\x01auth=Bearer <token>\x01\x01".
func parseXoauth2(raw string) (user, token string, err error) {
	trimmed, ok := strings.CutSuffix(raw, "\x01\x01")
	if !ok {
		return "", "", errors.New("initial response is not terminated with \\x01\\x01")
	}

	parts := strings.Split(trimmed, "\x01")
	if len(parts) != 2 {
		return "", "", errors.New("initial response does not have exactly 2 fields")
	}

	user, ok = strings.CutPrefix(parts[0], "user=")
	if !ok {
		return "", "", errors.New("first field is not user=")
	}

	token, ok = strings.CutPrefix(parts[1], "auth=Bearer ")
	if !ok {
		return "", "", errors.New("second field is not auth=Bearer ")
	}

	return user, token, nil
}

// startXoauth2Server runs an in-process IMAP server advertising AUTH=XOAUTH2.
func startXoauth2Server(t *testing.T, mech *xoauth2Server) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	srv := server.New(memory.New())
	// The memory backend only knows LOGIN; register XOAUTH2 alongside it.
	srv.EnableAuth("XOAUTH2", func(_ server.Conn) sasl.Server { return mech })
	// go-imap hides AUTHENTICATE behind STARTTLS unless plaintext auth is allowed.
	srv.AllowInsecureAuth = true

	//nolint:errcheck
	go srv.Serve(listener)
	t.Cleanup(func() { _ = srv.Close() })

	return listener.Addr().String()
}

func TestAuthenticateXoauth2OverIMAP(t *testing.T) {
	mech := &xoauth2Server{wantUser: "user@example.com", wantToken: "ya29.access-token"}
	addr := startXoauth2Server(t, mech)

	defer withConfig(YamlConfig{Auth: authOAuth2, User: "user@example.com"})()

	c, err := client.Dial(addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Logout() //nolint:errcheck

	if err := Authenticate(c, "ya29.access-token"); err != nil {
		t.Fatalf("Authenticate() error = %v (server saw %q)", err, mech.gotRaw)
	}

	// Guard the exact wire format, so a change here has to be deliberate.
	want := "user=user@example.com\x01auth=Bearer ya29.access-token\x01\x01"
	if mech.gotRaw != want {
		t.Fatalf("server received %q, want %q", mech.gotRaw, want)
	}
}

func TestAuthenticateXoauth2RejectsBadToken(t *testing.T) {
	mech := &xoauth2Server{wantUser: "user@example.com", wantToken: "good-token"}
	addr := startXoauth2Server(t, mech)

	defer withConfig(YamlConfig{Auth: authOAuth2, User: "user@example.com"})()

	c, err := client.Dial(addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Logout() //nolint:errcheck

	if err := Authenticate(c, "expired-token"); err == nil {
		t.Fatal("Authenticate() = nil, want an error for a rejected token")
	}
}

// Password configs must keep using LOGIN, not XOAUTH2.
func TestAuthenticateUsesLoginForPasswordAuth(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := server.New(memory.New())
	srv.AllowInsecureAuth = true
	//nolint:errcheck
	go srv.Serve(listener)
	t.Cleanup(func() { _ = srv.Close() })

	// Credentials the memory backend is seeded with.
	defer withConfig(YamlConfig{Auth: authPassword, User: "username", Pass: "password"})()

	c, err := client.Dial(listener.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Logout() //nolint:errcheck

	if err := Authenticate(c, ""); err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
}

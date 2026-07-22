package lib

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyPassFile(t *testing.T) {
	dir := t.TempDir()
	passPath := filepath.Join(dir, "secret.pass")
	if err := os.WriteFile(passPath, []byte("  secret-value \n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := YamlConfig{PassFile: passPath}
	if err := ApplyPassFile(&cfg); err != nil {
		t.Fatalf("ApplyPassFile() error = %v", err)
	}
	if cfg.Pass != "secret-value" {
		t.Fatalf("Pass = %q, want secret-value", cfg.Pass)
	}
}

func TestApplyPassFileMissing(t *testing.T) {
	cfg := YamlConfig{PassFile: filepath.Join(t.TempDir(), "missing.pass")}
	if err := ApplyPassFile(&cfg); err == nil {
		t.Fatal("ApplyPassFile() expected error for missing file")
	}
}

func TestApplyPassFileOverridesInlinePass(t *testing.T) {
	dir := t.TempDir()
	passPath := filepath.Join(dir, "secret.pass")
	if err := os.WriteFile(passPath, []byte("from-file"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := YamlConfig{Pass: "inline-pass", PassFile: passPath}
	if err := ApplyPassFile(&cfg); err != nil {
		t.Fatalf("ApplyPassFile() error = %v", err)
	}
	if cfg.Pass != "from-file" {
		t.Fatalf("Pass = %q, want from-file", cfg.Pass)
	}
}

func TestMaskSecrets(t *testing.T) {
	cfg := YamlConfig{
		Pass:     "real-password",
		PassFile: "/secret/path.pass",
	}
	MaskSecrets(&cfg)
	if cfg.Pass != secretMask {
		t.Fatalf("Pass = %q, want %q", cfg.Pass, secretMask)
	}
	if cfg.PassFile != secretMask {
		t.Fatalf("PassFile = %q, want %q", cfg.PassFile, secretMask)
	}
}

func TestMaskSecretsLeavesEmptyPassFile(t *testing.T) {
	cfg := YamlConfig{Pass: "real-password"}
	MaskSecrets(&cfg)
	if cfg.Pass != secretMask {
		t.Fatalf("Pass = %q, want %q", cfg.Pass, secretMask)
	}
	if cfg.PassFile != "" {
		t.Fatalf("PassFile = %q, want empty", cfg.PassFile)
	}
}

func TestMaskSecretsRedactsOAuthClientSecret(t *testing.T) {
	cfg := YamlConfig{OAuthClientID: "public-id", OAuthClientSecret: "real-secret"}
	MaskSecrets(&cfg)
	if cfg.OAuthClientSecret != secretMask {
		t.Fatalf("OAuthClientSecret = %q, want %q", cfg.OAuthClientSecret, secretMask)
	}
	// The client ID is not a secret, so it should stay readable.
	if cfg.OAuthClientID != "public-id" {
		t.Fatalf("OAuthClientID = %q, want public-id", cfg.OAuthClientID)
	}
}

func TestValidateAuth(t *testing.T) {
	base := YamlConfig{Host: "imap.example.com", User: "user@example.com"}

	withPass := base
	withPass.Pass = "secret"

	oauth := base
	oauth.Auth = "oauth2"
	oauth.OAuthClientID = "id"
	oauth.OAuthClientSecret = "secret"
	oauth.OAuthTokenFile = "/tmp/token.json"

	oauthNoSecret := oauth
	oauthNoSecret.OAuthClientSecret = ""

	oauthNoTokenFile := oauth
	oauthNoTokenFile.OAuthTokenFile = ""

	mixedCase := oauth
	mixedCase.Auth = " OAuth2 "

	tests := []struct {
		name    string
		cfg     YamlConfig
		wantErr bool
	}{
		{"password by default", withPass, false},
		{"password with no pass", base, true},
		{"oauth2 needs no password", oauth, false},
		{"oauth2 missing client secret", oauthNoSecret, true},
		{"oauth2 missing token file", oauthNoTokenFile, true},
		{"auth is case & space insensitive", mixedCase, false},
		{"unknown auth", YamlConfig{Host: "h", User: "u", Pass: "p", Auth: "kerberos"}, true},
		{"missing host", YamlConfig{User: "u", Pass: "p"}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.cfg
			err := ValidateAuth(&cfg)
			if tc.wantErr && err == nil {
				t.Fatal("ValidateAuth() = nil, want an error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("ValidateAuth() error = %v", err)
			}
		})
	}
}

func TestValidateAuthDefaultsToPassword(t *testing.T) {
	cfg := YamlConfig{Host: "imap.example.com", User: "user@example.com", Pass: "secret"}
	if err := ValidateAuth(&cfg); err != nil {
		t.Fatalf("ValidateAuth() error = %v", err)
	}
	if cfg.Auth != authPassword {
		t.Fatalf("Auth = %q, want %q", cfg.Auth, authPassword)
	}
	if cfg.UseOAuth2() {
		t.Fatal("UseOAuth2() = true, want false for the default auth")
	}
}

func TestUseOAuth2(t *testing.T) {
	if !(YamlConfig{Auth: "oauth2"}).UseOAuth2() {
		t.Fatal("UseOAuth2() = false for auth: oauth2")
	}
	if (YamlConfig{Auth: "password"}).UseOAuth2() {
		t.Fatal("UseOAuth2() = true for auth: password")
	}
	if (YamlConfig{}).UseOAuth2() {
		t.Fatal("UseOAuth2() = true for an unset auth")
	}
}

func TestRuleKeepSignaturesDefaultsToTrue(t *testing.T) {
	r := Rule{}
	if !r.KeepSignatures() {
		t.Fatal("KeepSignatures() = false, want true when PreserveSMIME is unset")
	}

	no := false
	r.PreserveSMIME = &no
	if r.KeepSignatures() {
		t.Fatal("KeepSignatures() = true, want false when PreserveSMIME is explicitly false")
	}

	yes := true
	r.PreserveSMIME = &yes
	if !r.KeepSignatures() {
		t.Fatal("KeepSignatures() = false, want true when PreserveSMIME is explicitly true")
	}
}

func TestReadConfigDefaultsKeepSignaturesToTrue(t *testing.T) {
	orig := Config
	t.Cleanup(func() { Config = orig })

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yml")
	yamlData := "host: imap.example.com\n" +
		"user: user@example.com\n" +
		"pass: secret\n" +
		"rules:\n" +
		"  - mailbox: INBOX\n" +
		"    actions: remove_attachments\n"
	if err := os.WriteFile(configPath, []byte(yamlData), 0o600); err != nil {
		t.Fatal(err)
	}

	ReadConfig(configPath)

	if len(Config.Rules) != 1 {
		t.Fatalf("Rules len = %d, want 1", len(Config.Rules))
	}
	if !Config.Rules[0].KeepSignatures() {
		t.Fatal("KeepSignatures() = false, want true by default after ReadConfig")
	}
}

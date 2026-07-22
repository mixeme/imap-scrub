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

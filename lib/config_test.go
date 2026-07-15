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

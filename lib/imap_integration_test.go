//go:build integration

package lib

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"gopkg.in/yaml.v3"
)

type imapTestConfig struct {
	Host     string `yaml:"host"`
	SSL      *bool  `yaml:"ssl"`
	Port     *int   `yaml:"port"`
	User     string `yaml:"user"`
	Pass     string `yaml:"pass"`
	PassFile string `yaml:"pass_file"`
}

func loadIntegrationConfig(t *testing.T) imapTestConfig {
	t.Helper()

	configPath := os.Getenv("IMAP_TEST_CONFIG")
	if configPath == "" {
		configPath = filepath.Join("dev", "imap-test.yml")
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Skipf("integration config not found: %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	var cfg imapTestConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse config: %v", err)
	}
	if cfg.PassFile != "" {
		passCfg := YamlConfig{PassFile: cfg.PassFile}
		if err := ApplyPassFile(&passCfg); err != nil {
			t.Fatalf("load pass_file: %v", err)
		}
		cfg.Pass = passCfg.Pass
	}
	if cfg.Host == "" || cfg.User == "" || cfg.Pass == "" {
		t.Fatal("integration config missing host, user, or password")
	}
	if cfg.SSL == nil {
		ssl := true
		cfg.SSL = &ssl
	}
	if cfg.Port == nil {
		port := 993
		if !*cfg.SSL {
			port = 143
		}
		cfg.Port = &port
	}
	return cfg
}

func connectIntegration(t *testing.T) *client.Client {
	t.Helper()
	cfg := loadIntegrationConfig(t)
	addr := cfg.Host + ":" + strconv.Itoa(*cfg.Port)

	var c *client.Client
	var err error
	if *cfg.SSL {
		c, err = client.DialTLS(addr, nil)
	} else {
		c, err = client.Dial(addr)
	}
	if err != nil {
		t.Fatalf("dial imap: %v", err)
	}
	if err := c.Login(cfg.User, cfg.Pass); err != nil {
		_ = c.Logout()
		t.Fatalf("login: %v", err)
	}
	t.Cleanup(func() { _ = c.Logout() })
	return c
}

func TestIntegrationFixturesPresent(t *testing.T) {
	c := connectIntegration(t)
	if _, err := c.Select("INBOX", true); err != nil {
		t.Fatalf("select INBOX: %v", err)
	}

	crit := imap.NewSearchCriteria()
	crit.Header.Set("Subject", "[imap-scrub-test]")
	ids, err := c.UidSearch(crit)
	if err != nil {
		t.Fatalf("search fixtures: %v", err)
	}
	if len(ids) < 5 {
		t.Fatalf("expected at least 5 seeded fixtures, got %d (run scripts/seed-imap-test.py)", len(ids))
	}
}

func TestIntegrationFromSenders(t *testing.T) {
	c := connectIntegration(t)
	if _, err := c.Select("INBOX", true); err != nil {
		t.Fatalf("select INBOX: %v", err)
	}

	for _, sender := range []string{"sender-a@test.com", "sender-b@test.com"} {
		crit := imap.NewSearchCriteria()
		crit.Header.Set("From", sender)
		crit.Header.Set("Subject", "[imap-scrub-test]")
		ids, err := c.UidSearch(crit)
		if err != nil {
			t.Fatalf("search from %s: %v", sender, err)
		}
		if len(ids) == 0 {
			t.Fatalf("expected messages from %s", sender)
		}
	}
}

func TestIntegrationOlderThanInternalDate(t *testing.T) {
	c := connectIntegration(t)
	if _, err := c.Select("INBOX", true); err != nil {
		t.Fatalf("select INBOX: %v", err)
	}

	crit := imap.NewSearchCriteria()
	crit.Header.Set("Subject", "[imap-scrub-test] old message from sender-a with attachment")
	ids, err := c.UidSearch(crit)
	if err != nil {
		t.Fatalf("search old fixture: %v", err)
	}
	if len(ids) == 0 {
		t.Fatal("old fixture message not found; run scripts/seed-imap-test.py")
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(ids[0])
	messages := make(chan *imap.Message, 1)
	if err := c.UidFetch(seqset, []imap.FetchItem{imap.FetchInternalDate, imap.FetchEnvelope}, messages); err != nil {
		t.Fatalf("fetch old fixture: %v", err)
	}
	msg := <-messages
	if msg == nil {
		t.Fatal("old fixture message not returned")
	}

	cutoff := BeginningOfDay(time.Now().AddDate(0, 0, -30))
	if !MessageIsOlderThan(msg, cutoff) {
		t.Fatalf("old fixture should be older than 30-day cutoff (internal=%s)", msg.InternalDate)
	}
}

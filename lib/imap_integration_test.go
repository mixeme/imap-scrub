//go:build integration

package lib

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
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

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
}

func loadIntegrationConfig(t *testing.T) imapTestConfig {
	t.Helper()
	root := repoRoot(t)

	configPath := os.Getenv("IMAP_TEST_CONFIG")
	if configPath == "" {
		configPath = filepath.Join(root, "dev", "imap-test.yml")
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
		passPath := cfg.PassFile
		if !filepath.IsAbs(passPath) {
			configDir := filepath.Dir(configPath)
			candidates := []string{
				filepath.Join(configDir, filepath.Base(passPath)), // same dir as config
				filepath.Join(filepath.Dir(configDir), passPath),  // repo root when config is in dev/
				filepath.Join(root, passPath),                     // this checkout's root
			}
			passPath = ""
			for _, cand := range candidates {
				if _, err := os.Stat(cand); err == nil {
					passPath = cand
					break
				}
			}
			if passPath == "" {
				t.Fatalf("pass_file %q not found (tried relative to config and repo roots)", cfg.PassFile)
			}
		}
		passCfg := YamlConfig{PassFile: passPath}
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

func fetchBySubject(t *testing.T, c *client.Client, subject string) *imap.Message {
	t.Helper()

	// Some IMAP servers tokenize unquoted HEADER Subject values on spaces
	// (e.g. treat "from" as a keyword). Search by fixture prefix, then match
	// the exact subject client-side.
	crit := imap.NewSearchCriteria()
	crit.Header = make(map[string][]string)
	crit.Header.Set("Subject", "[imap-scrub-test]")
	ids, err := c.UidSearch(crit)
	if err != nil {
		t.Fatalf("search fixtures: %v", err)
	}
	if len(ids) == 0 {
		t.Fatal("no fixtures found; run scripts/seed-imap-test.py")
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)
	messages := make(chan *imap.Message, len(ids))
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(seqset, []imap.FetchItem{imap.FetchInternalDate, imap.FetchEnvelope}, messages)
	}()

	var match *imap.Message
	for msg := range messages {
		if msg.Envelope != nil && msg.Envelope.Subject == subject {
			match = msg
		}
	}
	if err := <-done; err != nil {
		t.Fatalf("fetch fixtures: %v", err)
	}
	if match == nil {
		t.Fatalf("message not found: %q (run scripts/seed-imap-test.py)", subject)
	}
	return match
}

// fetchFullBySubject fetches a fixture with BODY.PEEK[] (same path as export_mailbox).
func fetchFullBySubject(t *testing.T, c *client.Client, subject string) *imap.Message {
	t.Helper()

	hdr := fetchBySubject(t, c, subject)

	seqset := new(imap.SeqSet)
	seqset.AddNum(hdr.Uid)
	var section imap.BodySectionName
	section.Peek = true
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchInternalDate, section.FetchItem()}

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(seqset, items, messages)
	}()

	msg := <-messages
	if err := <-done; err != nil {
		t.Fatalf("fetch full body: %v", err)
	}
	if msg == nil {
		t.Fatalf("full body not returned for %q", subject)
	}
	if len(msg.Body) == 0 {
		t.Fatalf("expected BODY.PEEK[] for %q", subject)
	}
	return msg
}

func TestIntegrationPassFileConnects(t *testing.T) {
	cfg := loadIntegrationConfig(t)
	if cfg.PassFile == "" && os.Getenv("IMAP_TEST_PASS") == "" {
		// CI may inject pass via env-built yaml without pass_file; still require login.
	}
	_ = connectIntegration(t)
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
	if len(ids) < 6 {
		t.Fatalf("expected at least 6 seeded fixtures, got %d (run scripts/seed-imap-test.py)", len(ids))
	}
}

func TestIntegrationFromSenders(t *testing.T) {
	c := connectIntegration(t)
	if _, err := c.Select("INBOX", true); err != nil {
		t.Fatalf("select INBOX: %v", err)
	}

	base := imap.SearchCriteria{}
	base.Header = make(map[string][]string)
	base.Header.Set("Subject", "[imap-scrub-test]")
	criteria, _ := FromSearchCriteria(base, "sender-a@test.com, sender-b@test.com")
	idGroups := make([][]uint32, 0, len(criteria))
	for _, crit := range criteria {
		ids, err := c.UidSearch(crit)
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		idGroups = append(idGroups, ids)
	}
	union := UnionUIDs(idGroups...)
	if len(union) < 2 {
		t.Fatalf("expected OR match across senders, got %d UIDs", len(union))
	}
}

func TestIntegrationOlderThanInternalDate(t *testing.T) {
	c := connectIntegration(t)
	if _, err := c.Select("INBOX", true); err != nil {
		t.Fatalf("select INBOX: %v", err)
	}

	cutoff := BeginningOfDay(time.Now().AddDate(0, 0, -30))

	old := fetchBySubject(t, c, "[imap-scrub-test] old message from sender-a with attachment")
	if !MessageIsOlderThan(old, cutoff) {
		t.Fatalf("old fixture should be older than 30-day cutoff (internal=%s)", old.InternalDate)
	}

	recent := fetchBySubject(t, c, "[imap-scrub-test] recent message from sender-b")
	if MessageIsOlderThan(recent, cutoff) {
		t.Fatalf("recent fixture should NOT match older_than 30 days (internal=%s)", recent.InternalDate)
	}
}

func TestIntegrationOlderThanIgnoresEnvelopeDate(t *testing.T) {
	c := connectIntegration(t)
	if _, err := c.Select("INBOX", true); err != nil {
		t.Fatalf("select INBOX: %v", err)
	}

	// Fixture: old Date header, recent IMAP internal date — must NOT match older_than.
	msg := fetchBySubject(t, c, "[imap-scrub-test] recent internal with old envelope Date")
	cutoff := BeginningOfDay(time.Now().AddDate(0, 0, -30))
	if MessageIsOlderThan(msg, cutoff) {
		t.Fatalf(
			"expected skip: internal=%s envelope=%v cutoff=%s",
			msg.InternalDate,
			msg.Envelope,
			cutoff,
		)
	}
	if msg.Envelope == nil || msg.Envelope.Date.IsZero() {
		t.Fatal("fixture missing envelope Date")
	}
	if !msg.Envelope.Date.Before(cutoff) {
		t.Fatalf("fixture envelope Date should be older than cutoff for this scenario: %s", msg.Envelope.Date)
	}
	if msg.InternalDate.Before(cutoff) {
		t.Fatalf("fixture internal date should be on/after cutoff: %s", msg.InternalDate)
	}
}

func TestIntegrationNewerThanInternalDate(t *testing.T) {
	c := connectIntegration(t)
	if _, err := c.Select("INBOX", true); err != nil {
		t.Fatalf("select INBOX: %v", err)
	}

	cutoff := BeginningOfDay(time.Now().AddDate(0, 0, -30))

	recent := fetchBySubject(t, c, "[imap-scrub-test] recent message from sender-b")
	if !MessageIsNewerThan(recent, cutoff) {
		t.Fatalf("recent fixture should match newer_than 30 days (internal=%s)", recent.InternalDate)
	}

	old := fetchBySubject(t, c, "[imap-scrub-test] old message from sender-a with attachment")
	if MessageIsNewerThan(old, cutoff) {
		t.Fatalf("old fixture should NOT match newer_than 30 days (internal=%s)", old.InternalDate)
	}
}

func TestIntegrationNewerThanIgnoresEnvelopeDate(t *testing.T) {
	c := connectIntegration(t)
	if _, err := c.Select("INBOX", true); err != nil {
		t.Fatalf("select INBOX: %v", err)
	}

	// Fixture: old Date header, recent IMAP internal date — must match newer_than.
	msg := fetchBySubject(t, c, "[imap-scrub-test] recent internal with old envelope Date")
	cutoff := BeginningOfDay(time.Now().AddDate(0, 0, -30))
	if !MessageIsNewerThan(msg, cutoff) {
		t.Fatalf(
			"expected match: internal=%s envelope=%v cutoff=%s",
			msg.InternalDate,
			msg.Envelope,
			cutoff,
		)
	}
	if msg.Envelope == nil || msg.Envelope.Date.IsZero() {
		t.Fatal("fixture missing envelope Date")
	}
	if !msg.Envelope.Date.Before(cutoff) {
		t.Fatalf("fixture envelope Date should be older than cutoff for this scenario: %s", msg.Envelope.Date)
	}
}

func TestIntegrationExportMailbox(t *testing.T) {
	c := connectIntegration(t)
	if _, err := c.Select("INBOX", true); err != nil {
		t.Fatalf("select INBOX: %v", err)
	}

	subject := "[imap-scrub-test] recent message from sender-b"
	msg := fetchFullBySubject(t, c, subject)

	dir := t.TempDir()
	orig := Config.SavePath
	Config.SavePath = dir
	defer func() { Config.SavePath = orig }()

	mboxFile, err := CreateMBOX("INBOX")
	if err != nil {
		t.Fatalf("CreateMBOX: %v", err)
	}

	if err := ExportMessage(msg, mboxFile.Writer); err != nil {
		_ = mboxFile.Close()
		t.Fatalf("ExportMessage: %v", err)
	}
	if err := mboxFile.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	outPath := filepath.Join(dir, "INBOX", "mbox")
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read mbox: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "From ") {
		t.Fatalf("mbox missing From line:\n%s", content)
	}
	if !strings.Contains(content, subject) {
		t.Fatalf("mbox missing subject %q:\n%s", subject, content)
	}
	if !strings.Contains(content, "sender-b@test.com") {
		t.Fatalf("mbox missing sender-b address:\n%s", content)
	}
}

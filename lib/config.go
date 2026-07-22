// Package lib contains various libs used within the main app
package lib

import (
	"os"
	"path"
	"strings"

	"go.yaml.in/yaml/v3"
)

var (
	// Config module global
	Config = YamlConfig{}

	validActions = map[string]bool{
		"delete":             true,
		"save_attachments":   true,
		"remove_attachments": true,
		"export_mailbox":     true,
	}
)

// YamlConfig config struct
type YamlConfig struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	SSL      *bool  `yaml:"ssl"`
	Port     *int   `yaml:"port"`
	User     string `yaml:"user"`
	Pass     string `yaml:"pass"`
	PassFile string `yaml:"pass_file"`
	SavePath string `yaml:"save_path"`
	UseTrash bool   `yaml:"use_trash"`
	Rules    []Rule `yaml:"rules"`
}

// Rule struct
type Rule struct {
	Mailbox        string `yaml:"mailbox"`
	Size           uint32 `yaml:"min_size"`   // KB
	OlderThan      int    `yaml:"older_than"` // days
	NewerThan      int    `yaml:"newer_than"` // days
	From           string `yaml:"from"`
	To             string `yaml:"to"`
	Subject        string `yaml:"subject"`
	Body           string `yaml:"body"`
	Text           string `yaml:"text"`
	Actions        string `yaml:"actions"`
	IncludeUnread  bool   `yaml:"include_unread"`
	IncludeStarred bool   `yaml:"include_starred"`
	PreserveSMIME  *bool  `yaml:"keep_signatures"` // default true, see Rule.KeepSignatures()
}

// ReadConfig reads & parses the config into global config
func ReadConfig(file string) {
	file = path.Clean(file)
	// #nosec
	yamlData, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(yamlData, &Config)

	if err != nil {
		Log.ErrorF("Error parsing %s:\n\n%s\n\n", file, err)
		os.Exit(2)
	}

	if Config.PassFile != "" {
		if err := ApplyPassFile(&Config); err != nil {
			Log.Error(err.Error())
			os.Exit(2)
		}
	}

	if Config.User == "" || Config.Pass == "" || Config.Host == "" {
		Log.Error("Please ensure host, user & password (or pass_file) are set")
		os.Exit(2)
	}

	if Config.SSL == nil {
		ssl := true
		Config.SSL = &ssl
	}

	if Config.Port == nil {
		port := 143
		if *Config.SSL {
			port = 993
		}
		Config.Port = &port
	}

	// change kB to bytes
	for x, item := range Config.Rules {
		Config.Rules[x].Size = item.Size * 1024

		if item.PreserveSMIME == nil {
			keep := true
			Config.Rules[x].PreserveSMIME = &keep
		}

		if item.Mailbox == "" {
			Log.Error("You must specify a mailbox for every rule")
			os.Exit(2)
		}

		if item.Actions == "" {
			Log.Error("You must have at least one action per rule")
			os.Exit(2)
		}

		raw := strings.ToLower(item.Actions)
		rawActions := strings.Split(raw, ",")
		actions := []string{}
		for _, a := range rawActions {
			a = strings.TrimSpace(a)
			if _, ok := validActions[a]; !ok {
				Log.ErrorF("\"%s\" is not a valid action", a)
				os.Exit(2)
			}
			actions = append(actions, a)
		}
		Config.Rules[x].Actions = strings.Join(actions, ", ")

		if Config.Rules[x].Delete() && Config.Rules[x].RemoveAttachments() {
			Log.Error("Your rule cannot contain both remove_attachments and delete")
			os.Exit(2)
		}
	}
}

// ApplyPassFile loads the password from cfg.PassFile into cfg.Pass.
func ApplyPassFile(cfg *YamlConfig) error {
	passFile := path.Clean(cfg.PassFile)
	passData, err := os.ReadFile(passFile)
	if err != nil {
		return err
	}
	cfg.Pass = strings.TrimSpace(string(passData))
	return nil
}

const secretMask = "**********"

// MaskSecrets redacts password fields for -p / print-config output.
func MaskSecrets(cfg *YamlConfig) {
	cfg.Pass = secretMask
	if cfg.PassFile != "" {
		cfg.PassFile = secretMask
	}
}

// Delete returns whether a rule is set to delete messages
func (r Rule) Delete() bool {
	return strings.Contains(r.Actions, "delete")
}

// RemoveAttachments returns whether a rule is set to delete messages
func (r Rule) RemoveAttachments() bool {
	return strings.Contains(r.Actions, "remove_attachments")
}

// SaveAttachments returns whether a rule is set to save attachments
func (r Rule) SaveAttachments() bool {
	return strings.Contains(r.Actions, "save_attachments")
}

// ExportMailbox returns whether a rule is set to export matching messages to mbox
func (r Rule) ExportMailbox() bool {
	return strings.Contains(r.Actions, "export_mailbox")
}

// KeepSignatures returns whether S/MIME signed messages (RFC 8551: smime.p7m,
// smime.p7s, smime.p7z) should be left untouched by remove_attachments rather
// than having their signature parts stripped like ordinary attachments.
// Defaults to true.
func (r Rule) KeepSignatures() bool {
	return r.PreserveSMIME == nil || *r.PreserveSMIME
}

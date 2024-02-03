package ssh_client

import (
	"fmt"
	"os"
	"strings"

	"github.com/kevinburke/ssh_config"
)

// Note: this should really be part of ssh_config,
// TODO add lazy open similar to UserSettings
// Oh oh! Done on git but not released.

type Settings struct {
	cfg *ssh_config.Config
}

func OpenSSHConfig(filename string) (*Settings, error) {
	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", filename, err)
	}
	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", filename, err)
	}
	return &Settings{cfg: cfg}, nil
}

func (s *Settings) Get(alias, key string) string {
	val, err := s.GetStrict(alias, key)
	if err != nil {
		return ""
	}
	return val
}

func (s *Settings) GetAll(alias, key string) []string {
	val, _ := s.GetAllStrict(alias, key)
	return val
}

func (s *Settings) GetStrict(alias, key string) (string, error) {
	val, err := s.cfg.Get(alias, key)
	if err != nil || val != "" {
		return val, err
	}
	return ssh_config.Default(key), nil
}

func (s *Settings) GetAllStrict(alias, key string) ([]string, error) {
	val, err := s.cfg.GetAll(alias, key)
	if err != nil || val != nil {
		return val, err
	}
	if def := MultipleDefault(key); def != nil {
		return def, nil
	}
	return []string{}, nil
}

// FROM ssh_config/validators.go
// these identities are used for SSH protocol 2
var defaultProtocol2Identities = []string{
	"~/.ssh/id_dsa",
	"~/.ssh/id_ecdsa",
	"~/.ssh/id_ed25519",
	"~/.ssh/id_rsa",
}

// MultipleDefault returns the default values for the given keyword, for example// {"22"} if the keyword is "Port". MultipleDefault returns nil if the keyword
// has no default, or if the keyword is unknown. Keyword matching is
// case-insensitive.
//
// MultipleDefault return more than one element for "IdentityFile" keyword.
func MultipleDefault(key string) []string {
	if def := ssh_config.Default(key); def != "" {
		if strings.ToLower(key) == strings.ToLower("IdentityFile") {
			return append([]string{def}, defaultProtocol2Identities...)
		}
		return []string{def}
	}
	return nil
}

// UserSettings overrides ssh_config.UserSettings to handle MultipleDefault
type UserSettings struct {
	*ssh_config.UserSettings
}

var DefaultUserSettings = &UserSettings{ssh_config.DefaultUserSettings}

func (u *UserSettings) GetAll(alias, key string) []string {
	val, _ := u.GetAllStrict(alias, key)
	return val
}

func (u *UserSettings) GetAllStrict(alias, key string) ([]string, error) {
	val, err := u.UserSettings.GetAllStrict(alias, key)
	if strings.ToLower(key) == strings.ToLower("IdentityFile") && len(val) == 1 {
		// TODO should discriminate explicit values from default ones to not fail on trying to open those files if not found.
		// Ugly hack, could be wrong if alias has a single "~/.ssh/identity"
		if val[0] == "~/.ssh/identity" {
			return MultipleDefault(key), nil
		}
	}
	if err != nil || val != nil {
		return val, err
	}
	return []string{}, nil
}

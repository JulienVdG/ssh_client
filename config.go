package ssh_client

import (
	"fmt"
	"os"

	"github.com/kevinburke/ssh_config"
)

// Note: this should really be part of ssh_config,
// TODO add lazy open similar to UserSettings

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
	// TODO: IdentityFile has multiple default values that we should return.
	if def := ssh_config.Default(key); def != "" {
		return []string{def}, nil
	}
	return []string{}, nil
}

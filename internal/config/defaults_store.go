package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/openhoo/vibecontainer/internal/domain"
)

type DefaultsStore struct {
	path string
}

func NewDefaultsStore() *DefaultsStore {
	return &DefaultsStore{path: DefaultsPath()}
}

func (s *DefaultsStore) Load() (domain.Defaults, error) {
	defaults := domain.DefaultDefaults()
	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaults, nil
		}
		return defaults, err
	}
	if err := json.Unmarshal(b, &defaults); err != nil {
		return domain.Defaults{}, err
	}
	if !defaults.Provider.Valid() {
		defaults.Provider = domain.DefaultDefaults().Provider
	}
	if defaults.ReadOnlyPort == 0 {
		defaults.ReadOnlyPort = 7681
	}
	if defaults.InteractivePort == 0 {
		defaults.InteractivePort = 7682
	}
	if defaults.TmuxAccess == "" {
		defaults.TmuxAccess = "read"
	}
	return defaults, nil
}

func (s *DefaultsStore) Save(defaults domain.Defaults) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(defaults, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, append(b, '\n'), 0o600)
}

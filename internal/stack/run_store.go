package stack

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/openhoo/vibecontainer/internal/config"
	"github.com/openhoo/vibecontainer/internal/domain"
)

type RunStore struct{}

func NewRunStore() *RunStore { return &RunStore{} }

func (s *RunStore) Save(opts domain.CreateOptions) (domain.RunMetadata, error) {
	runDir := config.RunDir(opts.Name)
	if err := os.MkdirAll(runDir, 0o700); err != nil {
		return domain.RunMetadata{}, err
	}
	compose, image, err := ComposeYAML(opts)
	if err != nil {
		return domain.RunMetadata{}, err
	}
	if err := os.WriteFile(config.RunComposePath(opts.Name), compose, 0o600); err != nil {
		return domain.RunMetadata{}, err
	}
	if err := os.WriteFile(config.RunEnvPath(opts.Name), EnvFile(opts), 0o600); err != nil {
		return domain.RunMetadata{}, err
	}
	now := time.Now().UTC()
	meta := domain.RunMetadata{
		Name:      opts.Name,
		Workspace: opts.WorkspacePath,
		Provider:  opts.Provider,
		Image:     image,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.writeMeta(meta); err != nil {
		return domain.RunMetadata{}, err
	}
	return meta, nil
}

func (s *RunStore) Touch(name string) error {
	meta, err := s.Load(name)
	if err != nil {
		return err
	}
	meta.UpdatedAt = time.Now().UTC()
	return s.writeMeta(meta)
}

func (s *RunStore) Load(name string) (domain.RunMetadata, error) {
	b, err := os.ReadFile(config.RunMetadataPath(name))
	if err != nil {
		return domain.RunMetadata{}, err
	}
	var meta domain.RunMetadata
	if err := json.Unmarshal(b, &meta); err != nil {
		return domain.RunMetadata{}, err
	}
	return meta, nil
}

func (s *RunStore) Exists(name string) bool {
	_, err := os.Stat(config.RunDir(name))
	return err == nil
}

func (s *RunStore) List() ([]domain.RunMetadata, error) {
	if err := os.MkdirAll(config.RunsDir(), 0o700); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(config.RunsDir())
	if err != nil {
		return nil, err
	}
	out := make([]domain.RunMetadata, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		meta, err := s.Load(e.Name())
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		out = append(out, meta)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *RunStore) Delete(name string) error {
	return os.RemoveAll(config.RunDir(name))
}

func (s *RunStore) writeMeta(meta domain.RunMetadata) error {
	path := config.RunMetadataPath(meta.Name)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o600)
}

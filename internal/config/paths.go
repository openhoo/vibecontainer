package config

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

func ConfigDir() string {
	return filepath.Join(xdg.ConfigHome, "vibecontainer")
}

func DefaultsPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

func DataDir() string {
	return filepath.Join(xdg.DataHome, "vibecontainer")
}

func RunsDir() string {
	return filepath.Join(DataDir(), "runs")
}

func RunDir(name string) string {
	return filepath.Join(RunsDir(), name)
}

func RunComposePath(name string) string {
	return filepath.Join(RunDir(name), "compose.yaml")
}

func RunEnvPath(name string) string {
	return filepath.Join(RunDir(name), ".env")
}

func RunMetadataPath(name string) string {
	return filepath.Join(RunDir(name), "run.json")
}

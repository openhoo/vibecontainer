package app

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func requireStackName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("--name is required")
	}
	return nil
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("cmd", "/c", "start", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

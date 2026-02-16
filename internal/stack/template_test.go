package stack

import (
	"strings"
	"testing"

	"github.com/openhoo/vibecontainer/internal/domain"
)

func TestComposeYAMLIncludesCloudflaredAndPorts(t *testing.T) {
	opts := domain.CreateOptions{
		Name:            "demo-stack",
		WorkspacePath:   "/tmp/workspace",
		Provider:        domain.ProviderCodex,
		ReadOnlyPort:    9001,
		Interactive:     true,
		InteractivePort: 9002,
		FirewallEnable:  true,
		TunnelEnable:    true,
		Auth: domain.Auth{
			TunnelToken:  "abc",
			OpenAIAPIKey: "sk-123",
		},
	}
	b, image, err := ComposeYAML(opts)
	if err != nil {
		t.Fatalf("compose generation failed: %v", err)
	}
	if image == "" {
		t.Fatal("expected image")
	}
	s := string(b)
	if !strings.Contains(s, "cloudflared") {
		t.Fatal("expected cloudflared service")
	}
	if !strings.Contains(s, "127.0.0.1:9001:7681") || !strings.Contains(s, "127.0.0.1:9002:7682") {
		t.Fatal("expected mapped ports")
	}
	if !strings.Contains(s, "working_dir: /workspace") || !strings.Contains(s, "/tmp/workspace:/workspace") {
		t.Fatal("expected workspace mount and working_dir")
	}
}

func TestComposeYAMLOmitsCloudflaredWhenDisabled(t *testing.T) {
	opts := domain.CreateOptions{
		Name:            "demo-stack",
		Provider:        domain.ProviderCodex,
		ReadOnlyPort:    9001,
		Interactive:     false,
		InteractivePort: 9002,
		FirewallEnable:  true,
		TunnelEnable:    false,
		Auth: domain.Auth{
			OpenAIAPIKey: "sk-123",
		},
	}
	b, _, err := ComposeYAML(opts)
	if err != nil {
		t.Fatalf("compose generation failed: %v", err)
	}
	s := string(b)
	if strings.Contains(s, "cloudflared") {
		t.Fatal("expected no cloudflared service when tunnel is disabled")
	}
}

func TestEnvFileContainsTunnelToken(t *testing.T) {
	env := string(EnvFile(domain.CreateOptions{
		TunnelEnable: true,
		Auth:         domain.Auth{TunnelToken: "tok"},
	}))
	if !strings.Contains(env, "TUNNEL_TOKEN=") {
		t.Fatal("expected tunnel token in env")
	}
}

func TestEnvFileEmptyWhenTunnelDisabled(t *testing.T) {
	env := EnvFile(domain.CreateOptions{
		TunnelEnable: false,
		Auth:         domain.Auth{TunnelToken: "tok"},
	})
	if len(env) != 0 {
		t.Fatalf("expected empty env file when tunnel disabled, got %q", string(env))
	}
}

func TestComposeYAMLOmitsWorkspaceWhenUnset(t *testing.T) {
	opts := domain.CreateOptions{
		Name:            "demo-stack",
		Provider:        domain.ProviderBase,
		ReadOnlyPort:    9001,
		Interactive:     false,
		InteractivePort: 9002,
		FirewallEnable:  true,
		TunnelEnable:    true,
		Auth: domain.Auth{
			TunnelToken: "abc",
		},
	}
	b, _, err := ComposeYAML(opts)
	if err != nil {
		t.Fatalf("compose generation failed: %v", err)
	}
	s := string(b)
	if strings.Contains(s, "working_dir: /workspace") || strings.Contains(s, ":/workspace") {
		t.Fatal("expected no workspace mount by default")
	}
}

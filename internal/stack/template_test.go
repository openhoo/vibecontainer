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
		TmuxAccess:      "write",
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
	// Secrets must use env-var substitution, not literal values
	if strings.Contains(s, "sk-123") {
		t.Fatal("compose YAML must not contain literal secret values")
	}
	if !strings.Contains(s, "${OPENAI_API_KEY}") {
		t.Fatal("expected ${OPENAI_API_KEY} substitution in compose YAML")
	}
}

func TestComposeYAMLOmitsCloudflaredWhenDisabled(t *testing.T) {
	opts := domain.CreateOptions{
		Name:            "demo-stack",
		Provider:        domain.ProviderCodex,
		ReadOnlyPort:    9001,
		TmuxAccess:      "read",
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
	if strings.Contains(s, "sk-123") {
		t.Fatal("compose YAML must not contain literal secret values")
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

func TestEnvFileContainsAllSecrets(t *testing.T) {
	env := string(EnvFile(domain.CreateOptions{
		Provider:     domain.ProviderCodex,
		TunnelEnable: true,
		Auth: domain.Auth{
			TunnelToken:  "tunnel-tok",
			OpenAIAPIKey: "sk-123",
			CodexAPIKey:  "cx-456",
		},
	}))
	for _, want := range []string{"TUNNEL_TOKEN=tunnel-tok", "OPENAI_API_KEY=sk-123", "CODEX_API_KEY=cx-456"} {
		if !strings.Contains(env, want) {
			t.Fatalf("expected env file to contain %q, got:\n%s", want, env)
		}
	}
}

func TestEnvFileContainsClaudeSecrets(t *testing.T) {
	env := string(EnvFile(domain.CreateOptions{
		Provider: domain.ProviderClaude,
		Auth: domain.Auth{
			ClaudeOAuthToken: "oauth-tok",
			AnthropicAPIKey:  "ant-key",
		},
	}))
	for _, want := range []string{"CLAUDE_CODE_OAUTH_TOKEN=oauth-tok", "ANTHROPIC_API_KEY=ant-key"} {
		if !strings.Contains(env, want) {
			t.Fatalf("expected env file to contain %q, got:\n%s", want, env)
		}
	}
}

func TestComposeYAMLOmitsWorkspaceWhenUnset(t *testing.T) {
	opts := domain.CreateOptions{
		Name:            "demo-stack",
		Provider:        domain.ProviderBase,
		ReadOnlyPort:    9001,
		TmuxAccess:      "read",
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

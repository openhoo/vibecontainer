package stack

import (
	"fmt"
	"sort"
	"strings"

	"github.com/openhoo/vibecontainer/internal/domain"
	"gopkg.in/yaml.v3"
)

const (
	managedLabel  = "com.openhoo.vibecontainer.managed"
	stackLabel    = "com.openhoo.vibecontainer.stack"
	providerLabel = "com.openhoo.vibecontainer.provider"
	serviceLabel  = "com.openhoo.vibecontainer.service"
	versionLabel  = "com.openhoo.vibecontainer.version"
)

func DefaultImage(provider domain.Provider) string {
	switch provider {
	case domain.ProviderBase:
		return "ghcr.io/openhoo/vibecontainer:latest"
	case domain.ProviderClaude:
		return "ghcr.io/openhoo/vibecontainer:claude"
	case domain.ProviderCodex:
		return "ghcr.io/openhoo/vibecontainer:codex"
	default:
		return "ghcr.io/openhoo/vibecontainer:latest"
	}
}

type composeFile struct {
	Services map[string]service `yaml:"services"`
}

type service struct {
	Image       string            `yaml:"image"`
	Container   string            `yaml:"container_name,omitempty"`
	WorkingDir  string            `yaml:"working_dir,omitempty"`
	CapAdd      []string          `yaml:"cap_add,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	Ports       []string          `yaml:"ports,omitempty"`
	DependsOn   []string          `yaml:"depends_on,omitempty"`
	Command     string            `yaml:"command,omitempty"`
	NetworkMode string            `yaml:"network_mode,omitempty"`
	Restart     string            `yaml:"restart,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
}

func ComposeYAML(opts domain.CreateOptions) ([]byte, string, error) {
	image := opts.Image
	if strings.TrimSpace(image) == "" {
		image = DefaultImage(opts.Provider)
	}

	tmuxEnabled := opts.TmuxAccess == "read" || opts.TmuxAccess == "write"
	env := map[string]string{
		"TMUX_WEB_ENABLE":             boolTo01(tmuxEnabled),
		"TMUX_WEB_INTERACTIVE_ENABLE": boolTo01(opts.TmuxAccess == "write"),
		"FIREWALL_ENABLE":             boolTo01(opts.FirewallEnable),
	}
	if opts.TTYDCredential != "" {
		env["TTYD_CREDENTIAL"] = "${TTYD_CREDENTIAL}"
	}

	switch opts.Provider {
	case domain.ProviderClaude:
		if opts.Auth.ClaudeOAuthToken != "" {
			env["CLAUDE_CODE_OAUTH_TOKEN"] = "${CLAUDE_CODE_OAUTH_TOKEN}"
		}
		if opts.Auth.AnthropicAPIKey != "" {
			env["ANTHROPIC_API_KEY"] = "${ANTHROPIC_API_KEY}"
		}
	case domain.ProviderCodex:
		if opts.Auth.CodexAuthJSON != "" {
			env["CODEX_AUTH_JSON"] = "${CODEX_AUTH_JSON}"
		}
		if opts.Auth.OpenAIAPIKey != "" {
			env["OPENAI_API_KEY"] = "${OPENAI_API_KEY}"
		}
		if opts.Auth.CodexAPIKey != "" {
			env["CODEX_API_KEY"] = "${CODEX_API_KEY}"
		}
	}

	var ports []string
	if tmuxEnabled {
		ports = append(ports, fmt.Sprintf("127.0.0.1:%d:7681", opts.ReadOnlyPort))
	}
	if opts.TmuxAccess == "write" {
		ports = append(ports, fmt.Sprintf("127.0.0.1:%d:7682", opts.InteractivePort))
	}

	labelsVibe := commonLabels(opts, "vibecontainer")
	vibeService := service{
		Image:       image,
		Container:   opts.Name + "-vibecontainer",
		CapAdd:      []string{"NET_ADMIN", "NET_RAW"},
		Environment: env,
		Ports:       ports,
		Restart:     "unless-stopped",
		Labels:      labelsVibe,
	}
	if strings.TrimSpace(opts.WorkspacePath) != "" {
		vibeService.WorkingDir = "/workspace"
		vibeService.Volumes = []string{fmt.Sprintf("%s:/workspace", opts.WorkspacePath)}
	}

	services := map[string]service{
		"vibecontainer": vibeService,
	}
	if opts.TunnelEnable {
		services["cloudflared"] = service{
			Image:       "cloudflare/cloudflared:2026.2.0",
			Container:   opts.Name + "-cloudflared",
			Command:     "tunnel run",
			Environment: map[string]string{"TUNNEL_TOKEN": "${TUNNEL_TOKEN}"},
			DependsOn:   []string{"vibecontainer"},
			NetworkMode: "service:vibecontainer",
			Restart:     "unless-stopped",
			Labels:      commonLabels(opts, "cloudflared"),
		}
	}
	compose := composeFile{Services: services}

	b, err := yaml.Marshal(compose)
	if err != nil {
		return nil, "", err
	}
	return b, image, nil
}

func EnvFile(opts domain.CreateOptions) []byte {
	env := map[string]string{}
	if opts.TunnelEnable && strings.TrimSpace(opts.Auth.TunnelToken) != "" {
		env["TUNNEL_TOKEN"] = opts.Auth.TunnelToken
	}
	if opts.TTYDCredential != "" {
		env["TTYD_CREDENTIAL"] = opts.TTYDCredential
	}
	switch opts.Provider {
	case domain.ProviderClaude:
		if opts.Auth.ClaudeOAuthToken != "" {
			env["CLAUDE_CODE_OAUTH_TOKEN"] = opts.Auth.ClaudeOAuthToken
		}
		if opts.Auth.AnthropicAPIKey != "" {
			env["ANTHROPIC_API_KEY"] = opts.Auth.AnthropicAPIKey
		}
	case domain.ProviderCodex:
		if opts.Auth.CodexAuthJSON != "" {
			env["CODEX_AUTH_JSON"] = opts.Auth.CodexAuthJSON
		}
		if opts.Auth.OpenAIAPIKey != "" {
			env["OPENAI_API_KEY"] = opts.Auth.OpenAIAPIKey
		}
		if opts.Auth.CodexAPIKey != "" {
			env["CODEX_API_KEY"] = opts.Auth.CodexAPIKey
		}
	}
	if len(env) == 0 {
		return []byte{}
	}
	lines := make([]string, 0, len(env))
	for k, v := range env {
		lines = append(lines, fmt.Sprintf("%s=%s", k, shellEscape(v)))
	}
	sort.Strings(lines)
	return []byte(strings.Join(lines, "\n") + "\n")
}

func commonLabels(opts domain.CreateOptions, serviceName string) map[string]string {
	return map[string]string{
		managedLabel:  "true",
		stackLabel:    opts.Name,
		providerLabel: string(opts.Provider),
		serviceLabel:  serviceName,
		versionLabel:  "1",
	}
}

func shellEscape(v string) string {
	if v == "" {
		return "''"
	}
	if strings.ContainsAny(v, " \t\n\"'$") {
		return "'" + strings.ReplaceAll(v, "'", "'\"'\"'") + "'"
	}
	return v
}

func boolTo01(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

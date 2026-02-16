package domain

import "time"

type Provider string

const (
	ProviderBase   Provider = "base"
	ProviderClaude Provider = "claude"
	ProviderCodex  Provider = "codex"
)

func (p Provider) Valid() bool {
	switch p {
	case ProviderBase, ProviderClaude, ProviderCodex:
		return true
	default:
		return false
	}
}

type Auth struct {
	ClaudeOAuthToken string `json:"-"`
	AnthropicAPIKey  string `json:"-"`
	CodexAuthJSON    string `json:"-"`
	OpenAIAPIKey     string `json:"-"`
	CodexAPIKey      string `json:"-"`
	TunnelToken      string `json:"-"`
}

type CreateOptions struct {
	Name            string   `json:"name"`
	WorkspacePath   string   `json:"workspace_path"`
	Provider        Provider `json:"provider"`
	Image           string   `json:"image,omitempty"`
	ReadOnlyPort    int      `json:"read_only_port"`
	Interactive     bool     `json:"interactive"`
	InteractivePort int      `json:"interactive_port"`
	TTYDCredential  string   `json:"ttyd_credential,omitempty"`
	FirewallEnable  bool     `json:"firewall_enable"`
	TunnelEnable    bool     `json:"tunnel_enable"`
	Auth            Auth     `json:"-"`
}

type RunMetadata struct {
	Name      string    `json:"name"`
	Workspace string    `json:"workspace"`
	Provider  Provider  `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Image     string    `json:"image"`
}

type Defaults struct {
	Provider        Provider `json:"provider"`
	ReadOnlyPort    int      `json:"read_only_port"`
	Interactive     bool     `json:"interactive"`
	InteractivePort int      `json:"interactive_port"`
	FirewallEnable  bool     `json:"firewall_enable"`
	TunnelEnable    bool     `json:"tunnel_enable"`
}

type ServiceStatus struct {
	Name    string `json:"Name"`
	State   string `json:"State"`
	Health  string `json:"Health"`
	Project string `json:"Project"`
}

func DefaultDefaults() Defaults {
	return Defaults{
		Provider:        ProviderCodex,
		ReadOnlyPort:    7681,
		Interactive:     false,
		InteractivePort: 7682,
		FirewallEnable:  true,
		TunnelEnable:    true,
	}
}

package validate

import (
	"testing"

	"github.com/openhoo/vibecontainer/internal/domain"
)

func TestCreateOptionsCodexValid(t *testing.T) {
	opts := domain.CreateOptions{
		Name:            "demo-stack",
		WorkspacePath:   t.TempDir(),
		Provider:        domain.ProviderCodex,
		ReadOnlyPort:    7681,
		TmuxAccess:      "read",
		InteractivePort: 7682,
		FirewallEnable:  true,
		TunnelEnable:    true,
		Auth: domain.Auth{
			TunnelToken:  "token",
			OpenAIAPIKey: "sk-123",
		},
	}
	if err := CreateOptions(opts); err != nil {
		t.Fatalf("expected valid options, got %v", err)
	}
}

func TestCreateOptionsClaudeMissingAuth(t *testing.T) {
	opts := domain.CreateOptions{
		Name:            "demo-stack",
		WorkspacePath:   t.TempDir(),
		Provider:        domain.ProviderClaude,
		ReadOnlyPort:    7681,
		TmuxAccess:      "read",
		InteractivePort: 7682,
		FirewallEnable:  true,
		TunnelEnable:    true,
		Auth:            domain.Auth{TunnelToken: "token"},
	}
	if err := CreateOptions(opts); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateCodexAuthJSON(t *testing.T) {
	payload := `{"auth_mode":"chatgptAuthTokens","tokens":{"id_token":"a","access_token":"b","refresh_token":"c"}}`
	if err := ValidateCodexAuthJSON(payload); err != nil {
		t.Fatalf("expected valid payload, got %v", err)
	}
}

func TestCreateOptionsAllowsNoWorkspace(t *testing.T) {
	opts := domain.CreateOptions{
		Name:            "demo-stack",
		Provider:        domain.ProviderCodex,
		ReadOnlyPort:    7681,
		TmuxAccess:      "read",
		InteractivePort: 7682,
		FirewallEnable:  true,
		TunnelEnable:    true,
		Auth: domain.Auth{
			TunnelToken:  "token",
			OpenAIAPIKey: "sk-123",
		},
	}
	if err := CreateOptions(opts); err != nil {
		t.Fatalf("expected valid options, got %v", err)
	}
}

func TestStackNameRejectsTrailingHyphen(t *testing.T) {
	opts := domain.CreateOptions{
		Name:         "my-stack-",
		Provider:     domain.ProviderCodex,
		ReadOnlyPort: 7681,
		TmuxAccess:   "none",
		Auth:         domain.Auth{OpenAIAPIKey: "sk-123"},
	}
	if err := CreateOptions(opts); err == nil {
		t.Fatal("expected error for trailing hyphen in name")
	}
}

func TestStackNameRejectsSingleChar(t *testing.T) {
	opts := domain.CreateOptions{
		Name:         "a",
		Provider:     domain.ProviderCodex,
		ReadOnlyPort: 7681,
		TmuxAccess:   "none",
		Auth:         domain.Auth{OpenAIAPIKey: "sk-123"},
	}
	if err := CreateOptions(opts); err == nil {
		t.Fatal("expected error for single-char name")
	}
}

func TestCreateOptionsValidWithoutTunnel(t *testing.T) {
	opts := domain.CreateOptions{
		Name:            "demo-stack",
		Provider:        domain.ProviderCodex,
		ReadOnlyPort:    7681,
		TmuxAccess:      "read",
		InteractivePort: 7682,
		FirewallEnable:  true,
		TunnelEnable:    false,
		Auth: domain.Auth{
			OpenAIAPIKey: "sk-123",
		},
	}
	if err := CreateOptions(opts); err != nil {
		t.Fatalf("expected valid options without tunnel, got %v", err)
	}
}

package keyring

import (
	"fmt"

	"github.com/openhoo/vibecontainer/internal/domain"
	"github.com/zalando/go-keyring"
)

const serviceName = "vibecontainer"

// Credential keys for the keyring
const (
	KeyClaudeOAuthToken = "claude_oauth_token"
	KeyAnthropicAPIKey  = "anthropic_api_key"
	KeyCodexAuthJSON    = "codex_auth_json"
	KeyOpenAIAPIKey     = "openai_api_key"
	KeyCodexAPIKey      = "codex_api_key"
	KeyTunnelToken      = "tunnel_token"
)

// Store provides secure credential storage using the system keychain
type Store struct {
	service string
}

// New creates a new keyring store
func New() *Store {
	return &Store{service: serviceName}
}

// SaveAuth saves all non-empty auth credentials to the keyring
func (s *Store) SaveAuth(auth domain.Auth) error {
	if auth.ClaudeOAuthToken != "" {
		if err := keyring.Set(s.service, KeyClaudeOAuthToken, auth.ClaudeOAuthToken); err != nil {
			return fmt.Errorf("save claude oauth token: %w", err)
		}
	}
	if auth.AnthropicAPIKey != "" {
		if err := keyring.Set(s.service, KeyAnthropicAPIKey, auth.AnthropicAPIKey); err != nil {
			return fmt.Errorf("save anthropic api key: %w", err)
		}
	}
	if auth.CodexAuthJSON != "" {
		if err := keyring.Set(s.service, KeyCodexAuthJSON, auth.CodexAuthJSON); err != nil {
			return fmt.Errorf("save codex auth json: %w", err)
		}
	}
	if auth.OpenAIAPIKey != "" {
		if err := keyring.Set(s.service, KeyOpenAIAPIKey, auth.OpenAIAPIKey); err != nil {
			return fmt.Errorf("save openai api key: %w", err)
		}
	}
	if auth.CodexAPIKey != "" {
		if err := keyring.Set(s.service, KeyCodexAPIKey, auth.CodexAPIKey); err != nil {
			return fmt.Errorf("save codex api key: %w", err)
		}
	}
	if auth.TunnelToken != "" {
		if err := keyring.Set(s.service, KeyTunnelToken, auth.TunnelToken); err != nil {
			return fmt.Errorf("save tunnel token: %w", err)
		}
	}
	return nil
}

// LoadAuth loads auth credentials from the keyring
// Returns a partially filled Auth struct with whatever credentials are available
func (s *Store) LoadAuth() domain.Auth {
	auth := domain.Auth{}

	// Try to load each credential, but don't fail if any are missing
	if val, err := keyring.Get(s.service, KeyClaudeOAuthToken); err == nil {
		auth.ClaudeOAuthToken = val
	}
	if val, err := keyring.Get(s.service, KeyAnthropicAPIKey); err == nil {
		auth.AnthropicAPIKey = val
	}
	if val, err := keyring.Get(s.service, KeyCodexAuthJSON); err == nil {
		auth.CodexAuthJSON = val
	}
	if val, err := keyring.Get(s.service, KeyOpenAIAPIKey); err == nil {
		auth.OpenAIAPIKey = val
	}
	if val, err := keyring.Get(s.service, KeyCodexAPIKey); err == nil {
		auth.CodexAPIKey = val
	}
	if val, err := keyring.Get(s.service, KeyTunnelToken); err == nil {
		auth.TunnelToken = val
	}

	return auth
}

// Get retrieves a single credential from the keyring
func (s *Store) Get(key string) (string, error) {
	return keyring.Get(s.service, key)
}

// Set stores a single credential in the keyring
func (s *Store) Set(key, value string) error {
	return keyring.Set(s.service, key, value)
}

// Delete removes a single credential from the keyring
func (s *Store) Delete(key string) error {
	return keyring.Delete(s.service, key)
}

// Clear removes all stored credentials from the keyring
func (s *Store) Clear() error {
	keys := []string{
		KeyClaudeOAuthToken,
		KeyAnthropicAPIKey,
		KeyCodexAuthJSON,
		KeyOpenAIAPIKey,
		KeyCodexAPIKey,
		KeyTunnelToken,
	}

	var firstErr error
	for _, key := range keys {
		if err := keyring.Delete(s.service, key); err != nil && err != keyring.ErrNotFound && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

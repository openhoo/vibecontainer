package validate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/openhoo/vibecontainer/internal/domain"
)

var stackNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,29}[a-z0-9]$`)

func CreateOptions(opts domain.CreateOptions) error {
	if !stackNameRe.MatchString(opts.Name) {
		return errors.New("name must be 2-31 chars, start and end with alphanumeric, and contain only lowercase alphanumeric or hyphens")
	}
	if !opts.Provider.Valid() {
		return errors.New("provider must be one of: base, claude, codex")
	}
	if strings.TrimSpace(opts.WorkspacePath) != "" {
		info, err := os.Stat(opts.WorkspacePath)
		if err != nil {
			return fmt.Errorf("workspace path is invalid: %w", err)
		}
		if !info.IsDir() {
			return errors.New("workspace path must be a directory")
		}
		if _, err := filepath.Abs(opts.WorkspacePath); err != nil {
			return fmt.Errorf("workspace path is invalid: %w", err)
		}
	}
	switch opts.TmuxAccess {
	case "none", "read", "write":
	default:
		return errors.New("tmux-access must be one of: none, read, write")
	}
	if opts.TmuxAccess == "read" || opts.TmuxAccess == "write" {
		if err := Port("readonly port", opts.ReadOnlyPort); err != nil {
			return err
		}
	}
	if opts.TmuxAccess == "write" {
		if err := Port("interactive port", opts.InteractivePort); err != nil {
			return err
		}
		if opts.ReadOnlyPort == opts.InteractivePort {
			return errors.New("readonly and interactive ports must differ")
		}
	}
	if opts.TTYDCredential != "" {
		if strings.ContainsAny(opts.TTYDCredential, " \t\n") || !strings.Contains(opts.TTYDCredential, ":") {
			return errors.New("ttyd credential must be in user:password format and contain no spaces")
		}
	}
	if opts.TunnelEnable {
		if strings.TrimSpace(opts.Auth.TunnelToken) == "" {
			return errors.New("tunnel token is required when tunnel is enabled")
		}
	}

	switch opts.Provider {
	case domain.ProviderClaude:
		if strings.TrimSpace(opts.Auth.ClaudeOAuthToken) == "" && strings.TrimSpace(opts.Auth.AnthropicAPIKey) == "" {
			return errors.New("claude requires CLAUDE_CODE_OAUTH_TOKEN or ANTHROPIC_API_KEY")
		}
	case domain.ProviderCodex:
		if strings.TrimSpace(opts.Auth.CodexAuthJSON) != "" {
			if err := ValidateCodexAuthJSON(opts.Auth.CodexAuthJSON); err != nil {
				return err
			}
		} else if strings.TrimSpace(opts.Auth.OpenAIAPIKey) == "" && strings.TrimSpace(opts.Auth.CodexAPIKey) == "" {
			return errors.New("codex requires CODEX_AUTH_JSON or OPENAI_API_KEY or CODEX_API_KEY")
		}
	}

	return nil
}

func ValidateCodexAuthJSON(payload string) error {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		return fmt.Errorf("codex auth json is invalid: %w", err)
	}
	if len(parsed) == 0 {
		return errors.New("codex auth json must be a non-empty object")
	}
	if authModeRaw, ok := parsed["auth_mode"]; ok {
		authMode, ok := authModeRaw.(string)
		if !ok {
			return errors.New("codex auth json auth_mode must be a string")
		}
		switch authMode {
		case "apikey", "chatgpt", "chatgptAuthTokens":
		default:
			return errors.New("codex auth json auth_mode must be one of: apikey, chatgpt, chatgptAuthTokens")
		}
	}
	if apiKey, ok := parsed["OPENAI_API_KEY"].(string); ok && strings.TrimSpace(apiKey) != "" {
		return nil
	}
	tokens, ok := parsed["tokens"].(map[string]any)
	if !ok {
		return errors.New("codex auth json must include OPENAI_API_KEY or tokens object")
	}
	for _, key := range []string{"id_token", "access_token", "refresh_token"} {
		if _, ok := tokens[key].(string); !ok {
			return fmt.Errorf("codex auth json tokens.%s must be a string", key)
		}
	}
	return nil
}

func Port(name string, val int) error {
	if val < 1 || val > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535", name)
	}
	return nil
}

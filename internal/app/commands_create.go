package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openhoo/vibecontainer/internal/config"
	"github.com/openhoo/vibecontainer/internal/docker"
	"github.com/openhoo/vibecontainer/internal/domain"
	"github.com/openhoo/vibecontainer/internal/keyring"
	"github.com/openhoo/vibecontainer/internal/stack"
	"github.com/openhoo/vibecontainer/internal/tui"
	"github.com/openhoo/vibecontainer/internal/validate"
	"github.com/spf13/cobra"
)

func newCreateCmd(defaults *config.DefaultsStore, runs *stack.RunStore, compose *docker.Compose) *cobra.Command {
	opts := domain.CreateOptions{}
	autoYes := false
	noSaveAuth := false

	cmd := &cobra.Command{
		Use:   "create [path]",
		Short: "Create and start a managed vibecontainer stack",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()

			def, err := defaults.Load()
			if err != nil {
				return fmt.Errorf("load defaults: %w", err)
			}
			opts = applyDefaults(cmd, opts, def)
			if len(args) == 1 {
				opts.WorkspacePath = args[0]
			}
			if strings.TrimSpace(opts.WorkspacePath) != "" {
				workspacePath, err := filepath.Abs(opts.WorkspacePath)
				if err != nil {
					return fmt.Errorf("resolve workspace path: %w", err)
				}
				opts.WorkspacePath = workspacePath
			}

			// Load stored credentials from keychain if not provided via flags
			kr := keyring.New()
			storedAuth := kr.LoadAuth()
			opts.Auth = mergeAuth(cmd, opts.Auth, storedAuth)

			if !autoYes {
				seedWorkspacePath := opts.WorkspacePath
				result, err := tui.RunCreateWizard(def, opts)
				if err != nil {
					return err
				}
				if !result.OK {
					return fmt.Errorf("create canceled")
				}
				opts = result.Options
				if opts.WorkspacePath == "" {
					opts.WorkspacePath = seedWorkspacePath
				}
			}

			opts.Provider = domain.Provider(strings.ToLower(strings.TrimSpace(string(opts.Provider))))
			if err := validate.CreateOptions(opts); err != nil {
				return err
			}
			if runs.Exists(opts.Name) {
				return fmt.Errorf("stack %q already exists", opts.Name)
			}

			meta, err := runs.Save(opts)
			if err != nil {
				return fmt.Errorf("save stack config: %w", err)
			}

			if err := compose.Up(ctx, opts.Name); err != nil {
				return err
			}
			if err := runs.Touch(opts.Name); err != nil {
				fmt.Fprintln(os.Stderr, "Warning: failed to update metadata:", err)
			}

			// Save credentials to keychain for next time
			if !noSaveAuth {
				if err := kr.SaveAuth(opts.Auth); err != nil {
					fmt.Fprintln(os.Stderr, "Warning: failed to save credentials to keychain:", err)
				}
			}

			if err := defaults.Save(domain.Defaults{
				Provider:        opts.Provider,
				ReadOnlyPort:    opts.ReadOnlyPort,
				Interactive:     opts.Interactive,
				InteractivePort: opts.InteractivePort,
				FirewallEnable:  opts.FirewallEnable,
				TunnelEnable:    opts.TunnelEnable,
			}); err != nil {
				fmt.Fprintln(os.Stderr, "Warning: failed to save defaults:", err)
			}

			fmt.Printf("Created stack %s (%s)\n", meta.Name, meta.Provider)
			fmt.Printf("Run dir: %s\n", config.RunDir(meta.Name))
			if opts.WorkspacePath == "" {
				fmt.Printf("Workspace: (not mapped)\n")
			} else {
				fmt.Printf("Workspace: %s\n", opts.WorkspacePath)
			}
			fmt.Printf("Read-only URL: http://127.0.0.1:%d\n", opts.ReadOnlyPort)
			if opts.Interactive {
				fmt.Printf("Interactive URL: http://127.0.0.1:%d\n", opts.InteractivePort)
			}
			if opts.TunnelEnable {
				fmt.Printf("Tunnel: enabled (Cloudflare)\n")
			} else {
				fmt.Printf("Tunnel: disabled\n")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&autoYes, "yes", false, "skip the TUI and use flags only")
	cmd.Flags().StringVar(&opts.Name, "name", "", "stack name")
	cmd.Flags().Var((*providerValue)(&opts.Provider), "provider", "provider: base|claude|codex")
	cmd.Flags().StringVar(&opts.Image, "image", "", "image override")
	cmd.Flags().IntVar(&opts.ReadOnlyPort, "readonly-port", 0, "read-only port")
	cmd.Flags().BoolVar(&opts.Interactive, "interactive", false, "enable interactive ttyd port")
	cmd.Flags().IntVar(&opts.InteractivePort, "interactive-port", 0, "interactive port")
	cmd.Flags().StringVar(&opts.TTYDCredential, "ttyd-credential", "", "ttyd basic auth credential user:password")
	cmd.Flags().BoolVar(&opts.FirewallEnable, "firewall-enable", true, "enable firewall inside container")
	cmd.Flags().BoolVar(&opts.TunnelEnable, "tunnel-enable", true, "enable cloudflare tunnel")
	cmd.Flags().StringVar(&opts.Auth.TunnelToken, "tunnel-token", "", "cloudflare tunnel token (required when tunnel is enabled)")
	cmd.Flags().StringVar(&opts.Auth.ClaudeOAuthToken, "claude-oauth-token", "", "claude oauth token")
	cmd.Flags().StringVar(&opts.Auth.AnthropicAPIKey, "anthropic-api-key", "", "anthropic api key")
	cmd.Flags().StringVar(&opts.Auth.CodexAuthJSON, "codex-auth-json", "", "codex auth json payload")
	cmd.Flags().StringVar(&opts.Auth.OpenAIAPIKey, "openai-api-key", "", "openai api key")
	cmd.Flags().StringVar(&opts.Auth.CodexAPIKey, "codex-api-key", "", "codex api key")

	return cmd
}

func applyDefaults(cmd *cobra.Command, opts domain.CreateOptions, d domain.Defaults) domain.CreateOptions {
	if !cmd.Flags().Changed("provider") || !opts.Provider.Valid() {
		opts.Provider = d.Provider
	}
	if !cmd.Flags().Changed("readonly-port") || opts.ReadOnlyPort == 0 {
		opts.ReadOnlyPort = d.ReadOnlyPort
	}
	if !cmd.Flags().Changed("interactive-port") || opts.InteractivePort == 0 {
		opts.InteractivePort = d.InteractivePort
	}
	if !cmd.Flags().Changed("interactive") {
		opts.Interactive = d.Interactive
	}
	if !cmd.Flags().Changed("firewall-enable") {
		opts.FirewallEnable = d.FirewallEnable
	}
	if !cmd.Flags().Changed("tunnel-enable") {
		opts.TunnelEnable = d.TunnelEnable
	}
	return opts
}

type providerValue domain.Provider

func (p *providerValue) String() string {
	if p == nil {
		return ""
	}
	return string(*p)
}

func (p *providerValue) Set(v string) error {
	value := strings.ToLower(strings.TrimSpace(v))
	pv := domain.Provider(value)
	if !pv.Valid() {
		return fmt.Errorf("invalid provider %q", v)
	}
	*p = providerValue(pv)
	return nil
}

func (p *providerValue) Type() string { return "provider" }

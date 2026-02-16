package app

import (
	"fmt"

	"github.com/openhoo/vibecontainer/internal/keyring"
	"github.com/spf13/cobra"
)

func newCredentialsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "credentials",
		Short: "Manage stored credentials",
		Long:  "Manage OAuth tokens and API keys stored in the system keychain",
	}

	cmd.AddCommand(newCredentialsClearCmd())
	cmd.AddCommand(newCredentialsListCmd())

	return cmd
}

func newCredentialsClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear all stored credentials from keychain",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr := keyring.New()
			if err := kr.Clear(); err != nil {
				return fmt.Errorf("clear credentials: %w", err)
			}
			fmt.Println("All credentials cleared from keychain")
			return nil
		},
	}
}

func newCredentialsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List which credentials are stored (without showing values)",
		RunE: func(cmd *cobra.Command, args []string) error {
			kr := keyring.New()
			auth := kr.LoadAuth()

			fmt.Println("Stored credentials:")
			printCredentialStatus("Claude OAuth Token", auth.ClaudeOAuthToken)
			printCredentialStatus("Anthropic API Key", auth.AnthropicAPIKey)
			printCredentialStatus("Codex Auth JSON", auth.CodexAuthJSON)
			printCredentialStatus("OpenAI API Key", auth.OpenAIAPIKey)
			printCredentialStatus("Codex API Key", auth.CodexAPIKey)
			printCredentialStatus("Tunnel Token", auth.TunnelToken)

			return nil
		},
	}
}

func printCredentialStatus(name, value string) {
	status := "✗ not stored"
	if value != "" {
		status = "✓ stored"
	}
	fmt.Printf("  %-24s %s\n", name+":", status)
}

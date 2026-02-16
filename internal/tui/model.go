package tui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/openhoo/vibecontainer/internal/domain"
)

type Result struct {
	Options domain.CreateOptions
	OK      bool
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginLeft(2).
			MarginTop(1).
			MarginBottom(1)

	reviewHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("39")).
				MarginLeft(2)

	reviewLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("111")).
				Width(22)

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	HeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Underline(true)
	ColumnStyle = lipgloss.NewStyle().PaddingRight(4)
)

func RunCreateWizard(defaults domain.Defaults, seed domain.CreateOptions) (Result, error) {
	opts := seed
	if !opts.Provider.Valid() {
		opts.Provider = defaults.Provider
	}
	if opts.ReadOnlyPort == 0 {
		opts.ReadOnlyPort = defaults.ReadOnlyPort
	}
	if opts.InteractivePort == 0 {
		opts.InteractivePort = defaults.InteractivePort
	}

	// Track whether saved credentials exist (static, for hide-funcs on confirms)
	hasClaudeOAuth := opts.Auth.ClaudeOAuthToken != ""
	hasAnthropicKey := opts.Auth.AnthropicAPIKey != ""
	hasOpenAIKey := opts.Auth.OpenAIAPIKey != ""
	hasCodexKey := opts.Auth.CodexAPIKey != ""
	hasCodexAuth := opts.Auth.CodexAuthJSON != ""
	hasTunnelToken := opts.Auth.TunnelToken != ""

	var (
		provider           = string(opts.Provider)
		tmuxExpose         = opts.TmuxAccess != "none"
		tmuxAccess         = opts.TmuxAccess
		firewall           = opts.FirewallEnable
		tunnelEnable       = opts.TunnelEnable
		readOnlyPortStr    = strconv.Itoa(opts.ReadOnlyPort)
		interactivePortStr = strconv.Itoa(opts.InteractivePort)
		customizeAdvanced  bool
		claudeAuthMethod   = "oauth"
		codexAuthMethod    = "openai"

		// Mutable: flipped by confirm groups; start true when saved value exists
		useExistingClaudeOAuth  = hasClaudeOAuth
		useExistingAnthropicKey = hasAnthropicKey
		useExistingOpenAIKey    = hasOpenAIKey
		useExistingCodexKey     = hasCodexKey
		useExistingCodexAuth    = hasCodexAuth
		useExistingTunnelToken  = hasTunnelToken

		// New credential inputs (empty; used only when user declines saved)
		newClaudeOAuth  string
		newAnthropicKey string
		newOpenAIKey    string
		newCodexKey     string
		newCodexAuth    string
		newTunnelToken  string
	)
	if tmuxAccess == "" {
		tmuxAccess = "read"
	}

	fmt.Println(titleStyle.Render("Vibecontainer Setup"))

	form := huh.NewForm(
		// Stack Name
		huh.NewGroup(
			huh.NewInput().
				Title("Stack Name").
				Description("A unique name for this container stack").
				Placeholder("my-stack").
				Value(&opts.Name).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("name is required")
					}
					return nil
				}),
		),

		// Provider
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Provider").
				Description("Choose the AI coding agent to run").
				Options(
					huh.NewOption("Codex (OpenAI)", "codex"),
					huh.NewOption("Claude (Anthropic)", "claude"),
					huh.NewOption("Base (Minimal)", "base"),
				).
				Value(&provider),
		),

		// Claude: auth method
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Claude Authentication").
				Description("Choose how to authenticate with Claude").
				Options(
					huh.NewOption("OAuth Token", "oauth"),
					huh.NewOption("API Key", "apikey"),
				).
				Value(&claudeAuthMethod),
		).WithHideFunc(func() bool { return provider != "claude" }),

		// Claude OAuth: use saved?
		huh.NewGroup(
			huh.NewConfirm().
				Title("Use saved Claude OAuth Token?").
				Value(&useExistingClaudeOAuth),
		).WithHideFunc(func() bool { return provider != "claude" || claudeAuthMethod != "oauth" || !hasClaudeOAuth }),

		// Claude OAuth: new input
		huh.NewGroup(
			huh.NewInput().
				Title("Claude OAuth Token").
				Password(true).
				Value(&newClaudeOAuth).
				Validate(notEmpty("oauth token is required")),
		).WithHideFunc(func() bool { return provider != "claude" || claudeAuthMethod != "oauth" || useExistingClaudeOAuth }),

		// Anthropic API Key: use saved?
		huh.NewGroup(
			huh.NewConfirm().
				Title("Use saved Anthropic API Key?").
				Value(&useExistingAnthropicKey),
		).WithHideFunc(func() bool { return provider != "claude" || claudeAuthMethod != "apikey" || !hasAnthropicKey }),

		// Anthropic API Key: new input
		huh.NewGroup(
			huh.NewInput().
				Title("Anthropic API Key").
				Password(true).
				Value(&newAnthropicKey).
				Validate(notEmpty("api key is required")),
		).WithHideFunc(func() bool { return provider != "claude" || claudeAuthMethod != "apikey" || useExistingAnthropicKey }),

		// Codex: auth method
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Codex Authentication").
				Description("Choose how to authenticate with Codex").
				Options(
					huh.NewOption("OpenAI API Key", "openai"),
					huh.NewOption("Codex API Key", "codex_key"),
					huh.NewOption("Auth JSON", "auth_json"),
				).
				Value(&codexAuthMethod),
		).WithHideFunc(func() bool { return provider != "codex" }),

		// OpenAI API Key: use saved?
		huh.NewGroup(
			huh.NewConfirm().
				Title("Use saved OpenAI API Key?").
				Value(&useExistingOpenAIKey),
		).WithHideFunc(func() bool { return provider != "codex" || codexAuthMethod != "openai" || !hasOpenAIKey }),

		// OpenAI API Key: new input
		huh.NewGroup(
			huh.NewInput().
				Title("OpenAI API Key").
				Password(true).
				Value(&newOpenAIKey).
				Validate(notEmpty("api key is required")),
		).WithHideFunc(func() bool { return provider != "codex" || codexAuthMethod != "openai" || useExistingOpenAIKey }),

		// Codex API Key: use saved?
		huh.NewGroup(
			huh.NewConfirm().
				Title("Use saved Codex API Key?").
				Value(&useExistingCodexKey),
		).WithHideFunc(func() bool { return provider != "codex" || codexAuthMethod != "codex_key" || !hasCodexKey }),

		// Codex API Key: new input
		huh.NewGroup(
			huh.NewInput().
				Title("Codex API Key").
				Password(true).
				Value(&newCodexKey).
				Validate(notEmpty("api key is required")),
		).WithHideFunc(func() bool { return provider != "codex" || codexAuthMethod != "codex_key" || useExistingCodexKey }),

		// Codex Auth JSON: use saved?
		huh.NewGroup(
			huh.NewConfirm().
				Title("Use saved Codex Auth JSON?").
				Value(&useExistingCodexAuth),
		).WithHideFunc(func() bool { return provider != "codex" || codexAuthMethod != "auth_json" || !hasCodexAuth }),

		// Codex Auth JSON: new input
		huh.NewGroup(
			huh.NewInput().
				Title("Codex Auth JSON").
				Password(true).
				Value(&newCodexAuth).
				Validate(notEmpty("auth json is required")),
		).WithHideFunc(func() bool { return provider != "codex" || codexAuthMethod != "auth_json" || useExistingCodexAuth }),

		// Workspace Path
		huh.NewGroup(
			huh.NewInput().
				Title("Workspace Path").
				Description("Local directory to mount into the container (leave empty to skip)").
				Placeholder("e.g. ./my-project").
				Value(&opts.WorkspacePath),
		),

		// Expose tmux
		huh.NewGroup(
			huh.NewConfirm().
				Title("Expose tmux?").
				Description("Make the tmux session accessible via a web browser").
				Value(&tmuxExpose),
		),

		// Tmux access level
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Tmux Access Level").
				Description("Choose the level of access for the web terminal").
				Options(
					huh.NewOption("Read-only (view only)", "read"),
					huh.NewOption("Read & Write (interactive)", "write"),
				).
				Value(&tmuxAccess),
		).WithHideFunc(func() bool { return !tmuxExpose }),

		// Cloudflare Tunnel
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable Cloudflare Tunnel?").
				Description("Expose your container via a Cloudflare tunnel for remote access").
				Value(&tunnelEnable),
		),

		// Tunnel Token: use saved?
		huh.NewGroup(
			huh.NewConfirm().
				Title("Use saved Cloudflare Tunnel Token?").
				Value(&useExistingTunnelToken),
		).WithHideFunc(func() bool { return !tunnelEnable || !hasTunnelToken }),

		// Tunnel Token: new input
		huh.NewGroup(
			huh.NewInput().
				Title("Cloudflare Tunnel Token").
				Password(true).
				Value(&newTunnelToken).
				Validate(notEmpty("tunnel token is required")),
		).WithHideFunc(func() bool { return !tunnelEnable || useExistingTunnelToken }),

		// Advanced settings gate
		huh.NewGroup(
			huh.NewConfirm().
				Title("Customize advanced settings?").
				Description("Ports, firewall, image overrides").
				Value(&customizeAdvanced),
		),

		// Read-only Port
		huh.NewGroup(
			huh.NewInput().
				Title("Read-only Port").
				Description("Port for the read-only terminal view").
				Value(&readOnlyPortStr).
				Validate(validatePort("read-only port")),
		).WithHideFunc(func() bool { return !customizeAdvanced }),

		// Interactive Port
		huh.NewGroup(
			huh.NewInput().
				Title("Interactive Port").
				Description("Port for the interactive terminal").
				Value(&interactivePortStr).
				Validate(validatePort("interactive port")),
		).WithHideFunc(func() bool { return !customizeAdvanced || tmuxAccess != "write" }),

		// Firewall
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable Firewall?").
				Description("Restrict outbound network access inside the container").
				Value(&firewall),
		).WithHideFunc(func() bool { return !customizeAdvanced }),

		// Image Override
		huh.NewGroup(
			huh.NewInput().
				Title("Image Override").
				Description("Custom Docker image (leave empty for default)").
				Placeholder("ghcr.io/openhoo/vibecontainer:latest").
				Value(&opts.Image),
		).WithHideFunc(func() bool { return !customizeAdvanced }),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return Result{OK: false}, nil
		}
		return Result{}, err
	}

	// Sync values back
	opts.Provider = domain.Provider(provider)
	if tmuxExpose {
		opts.TmuxAccess = tmuxAccess
	} else {
		opts.TmuxAccess = "none"
	}
	opts.FirewallEnable = firewall
	opts.TunnelEnable = tunnelEnable
	opts.ReadOnlyPort, _ = strconv.Atoi(readOnlyPortStr)
	opts.InteractivePort, _ = strconv.Atoi(interactivePortStr)

	// Sync credentials: use new value when user declined saved
	if !useExistingClaudeOAuth {
		opts.Auth.ClaudeOAuthToken = newClaudeOAuth
	}
	if !useExistingAnthropicKey {
		opts.Auth.AnthropicAPIKey = newAnthropicKey
	}
	if !useExistingOpenAIKey {
		opts.Auth.OpenAIAPIKey = newOpenAIKey
	}
	if !useExistingCodexKey {
		opts.Auth.CodexAPIKey = newCodexKey
	}
	if !useExistingCodexAuth {
		opts.Auth.CodexAuthJSON = newCodexAuth
	}
	if !useExistingTunnelToken {
		opts.Auth.TunnelToken = newTunnelToken
	}

	// Build auth description for review
	authDesc := authDescription(provider, claudeAuthMethod, codexAuthMethod)

	// Print review
	fmt.Println()
	fmt.Println(reviewHeaderStyle.Render("Configuration Review"))
	fmt.Println()
	printReview(opts, authDesc)
	fmt.Println()

	// Confirm
	ok, err := Confirm("Create this stack?", "", true)

	if err != nil {
		return Result{}, err
	}
	if !ok {
		return Result{OK: false}, nil
	}

	return Result{Options: opts, OK: true}, nil
}

func validatePort(name string) func(string) error {
	return func(s string) error {
		n, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("%s must be a number", name)
		}
		if n < 1 || n > 65535 {
			return fmt.Errorf("%s must be between 1 and 65535", name)
		}
		return nil
	}
}

func notEmpty(msg string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s", msg)
		}
		return nil
	}
}

func authDescription(provider, claudeMethod, codexMethod string) string {
	switch provider {
	case "claude":
		if claudeMethod == "oauth" {
			return "OAuth Token"
		}
		return "API Key"
	case "codex":
		switch codexMethod {
		case "openai":
			return "OpenAI API Key"
		case "codex_key":
			return "Codex API Key"
		case "auth_json":
			return "Auth JSON"
		}
	}
	return "none"
}

func printReview(opts domain.CreateOptions, authDesc string) {
	divider := "  " + dividerStyle.Render(strings.Repeat("─", 40))
	fmt.Println(divider)

	line := func(label, value string) {
		fmt.Printf("  %s %s\n", reviewLabelStyle.Render(label), value)
	}

	line("Stack Name:", opts.Name)
	line("Provider:", string(opts.Provider))
	line("Auth:", authDesc)
	if opts.WorkspacePath != "" {
		line("Workspace:", opts.WorkspacePath)
	} else {
		line("Workspace:", "(not mapped)")
	}
	line("Tunnel:", boolWord(opts.TunnelEnable))
	line("Tmux Access:", opts.TmuxAccess)
	if opts.TmuxAccess == "read" || opts.TmuxAccess == "write" {
		line("Read-only Port:", strconv.Itoa(opts.ReadOnlyPort))
	}
	if opts.TmuxAccess == "write" {
		line("Interactive Port:", strconv.Itoa(opts.InteractivePort))
	}
	line("Firewall:", boolWord(opts.FirewallEnable))
	if opts.Image != "" {
		line("Image:", opts.Image)
	}

	fmt.Println(divider)
}

func boolWord(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func Confirm(title string, description string, defaultYes bool) (bool, error) {
	confirm := defaultYes
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Description(description).
				Value(&confirm),
		),
	).WithTheme(huh.ThemeCharm()).Run()

	if err != nil {
		if err == huh.ErrUserAborted {
			return false, nil
		}
		return false, err
	}
	return confirm, nil
}

func RenderList(metas []domain.RunMetadata, stateByStack map[string][]string) {
	header := []string{"NAME", "PROVIDER", "STATE", "UPDATED"}
	rows := [][]string{}

	for _, m := range metas {
		states := stateByStack[m.Name]
		sort.Strings(states)
		state := "not-created"
		if len(states) > 0 {
			state = strings.Join(states, ",")
		}
		rows = append(rows, []string{m.Name, string(m.Provider), state, fmtTime(m.UpdatedAt)})
	}

	renderTable(header, rows)
}

func RenderStatus(statuses []domain.ServiceStatus) {
	header := []string{"SERVICE", "STATE", "HEALTH"}
	rows := [][]string{}

	for _, s := range statuses {
		health := strings.TrimSpace(s.Health)
		if health == "" {
			health = "-"
		}
		rows = append(rows, []string{s.Name, s.State, health})
	}

	renderTable(header, rows)
}

func renderTable(header []string, rows [][]string) {
	colWidths := make([]int, len(header))
	for i, h := range header {
		colWidths[i] = len(h)
	}

	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Render header
	for i, h := range header {
		fmt.Print(HeaderStyle.Width(colWidths[i] + 4).Render(h))
	}
	fmt.Println()

	// Render rows
	for _, row := range rows {
		for i, cell := range row {
			fmt.Print(lipgloss.NewStyle().Width(colWidths[i] + 4).Render(cell))
		}
		fmt.Println()
	}
}

func fmtTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

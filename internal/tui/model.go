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

	var (
		provider           = string(opts.Provider)
		interactive        = opts.Interactive
		firewall           = opts.FirewallEnable
		tunnelEnable       = opts.TunnelEnable
		readOnlyPortStr    = strconv.Itoa(opts.ReadOnlyPort)
		interactivePortStr = strconv.Itoa(opts.InteractivePort)
		customizeAdvanced  bool
		claudeAuthMethod   = "oauth"
		codexAuthMethod    = "openai"
	)

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

		// Claude: OAuth Token
		huh.NewGroup(
			huh.NewInput().
				Title("Claude OAuth Token").
				Password(true).
				Value(&opts.Auth.ClaudeOAuthToken).
				Validate(notEmpty("oauth token is required")),
		).WithHideFunc(func() bool { return provider != "claude" || claudeAuthMethod != "oauth" }),

		// Claude: API Key
		huh.NewGroup(
			huh.NewInput().
				Title("Anthropic API Key").
				Password(true).
				Value(&opts.Auth.AnthropicAPIKey).
				Validate(notEmpty("api key is required")),
		).WithHideFunc(func() bool { return provider != "claude" || claudeAuthMethod != "apikey" }),

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

		// Codex: OpenAI API Key
		huh.NewGroup(
			huh.NewInput().
				Title("OpenAI API Key").
				Password(true).
				Value(&opts.Auth.OpenAIAPIKey).
				Validate(notEmpty("api key is required")),
		).WithHideFunc(func() bool { return provider != "codex" || codexAuthMethod != "openai" }),

		// Codex: Codex API Key
		huh.NewGroup(
			huh.NewInput().
				Title("Codex API Key").
				Password(true).
				Value(&opts.Auth.CodexAPIKey).
				Validate(notEmpty("api key is required")),
		).WithHideFunc(func() bool { return provider != "codex" || codexAuthMethod != "codex_key" }),

		// Codex: Auth JSON
		huh.NewGroup(
			huh.NewInput().
				Title("Codex Auth JSON").
				Password(true).
				Value(&opts.Auth.CodexAuthJSON).
				Validate(notEmpty("auth json is required")),
		).WithHideFunc(func() bool { return provider != "codex" || codexAuthMethod != "auth_json" }),

		// Workspace Path
		huh.NewGroup(
			huh.NewInput().
				Title("Workspace Path").
				Description("Local directory to mount into the container (leave empty to skip)").
				Placeholder("e.g. ./my-project").
				Value(&opts.WorkspacePath),
		),

		// Cloudflare Tunnel
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable Cloudflare Tunnel?").
				Description("Expose your container via a Cloudflare tunnel for remote access").
				Value(&tunnelEnable),
		),

		// Tunnel Token
		huh.NewGroup(
			huh.NewInput().
				Title("Cloudflare Tunnel Token").
				Password(true).
				Value(&opts.Auth.TunnelToken).
				Validate(notEmpty("tunnel token is required")),
		).WithHideFunc(func() bool { return !tunnelEnable }),

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
				Value(&readOnlyPortStr),
		).WithHideFunc(func() bool { return !customizeAdvanced }),

		// Interactive TTY
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable Interactive TTY?").
				Description("Run an interactive terminal alongside the read-only view").
				Value(&interactive),
		).WithHideFunc(func() bool { return !customizeAdvanced }),

		// Interactive Port
		huh.NewGroup(
			huh.NewInput().
				Title("Interactive Port").
				Description("Port for the interactive terminal").
				Value(&interactivePortStr),
		).WithHideFunc(func() bool { return !customizeAdvanced || !interactive }),

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
	opts.Interactive = interactive
	opts.FirewallEnable = firewall
	opts.TunnelEnable = tunnelEnable
	opts.ReadOnlyPort, _ = strconv.Atoi(readOnlyPortStr)
	opts.InteractivePort, _ = strconv.Atoi(interactivePortStr)

	// Build auth description for review
	authDesc := authDescription(provider, claudeAuthMethod, codexAuthMethod)

	// Print review
	fmt.Println()
	fmt.Println(reviewHeaderStyle.Render("Configuration Review"))
	fmt.Println()
	printReview(opts, authDesc)
	fmt.Println()

	// Confirm
	ok, err := Confirm("Create this stack?", "")
	if err != nil {
		return Result{}, err
	}
	if !ok {
		return Result{OK: false}, nil
	}

	return Result{Options: opts, OK: true}, nil
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
	line("Read-only Port:", strconv.Itoa(opts.ReadOnlyPort))
	line("Interactive:", boolWord(opts.Interactive))
	if opts.Interactive {
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

func Confirm(title string, description string) (bool, error) {
	var confirm bool
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

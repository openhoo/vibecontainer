package app

import (
	"fmt"
	"os"

	"github.com/openhoo/vibecontainer/internal/config"
	"github.com/openhoo/vibecontainer/internal/docker"
	"github.com/openhoo/vibecontainer/internal/stack"
	"github.com/spf13/cobra"
)

type App struct {
	version string
	commit  string
	date    string
}

func New(version, commit, date string) *App {
	return &App{version: version, commit: commit, date: date}
}

func (a *App) Execute() error {
	store := config.NewDefaultsStore()
	runs := stack.NewRunStore()
	compose := docker.NewCompose(docker.NewExecRunner())

	root := &cobra.Command{
		Use:           "vibecontainer",
		Short:         "Manage vibecontainer stacks",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.Version = fmt.Sprintf("%s (commit=%s date=%s)", a.version, a.commit, a.date)
	root.SetVersionTemplate("{{.Version}}\n")

	root.AddCommand(newCreateCmd(store, runs, compose))
	root.AddCommand(newListCmd(runs, compose))
	root.AddCommand(newStatusCmd(runs, compose))
	root.AddCommand(newStartCmd(runs, compose))
	root.AddCommand(newStopCmd(runs, compose))
	root.AddCommand(newRestartCmd(runs, compose))
	root.AddCommand(newLogsCmd(runs, compose))
	root.AddCommand(newRemoveCmd(runs, compose))
	root.AddCommand(newCredentialsCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return err
	}
	return nil
}

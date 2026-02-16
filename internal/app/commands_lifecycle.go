package app

import (
	"context"
	"fmt"
	"time"

	"github.com/openhoo/vibecontainer/internal/docker"
	"github.com/openhoo/vibecontainer/internal/stack"
	"github.com/openhoo/vibecontainer/internal/tui"
	"github.com/spf13/cobra"
)

func newStartCmd(runs *stack.RunStore, compose *docker.Compose) *cobra.Command {
	name := ""
	cmd := &cobra.Command{
		Use:   "start --name <stack>",
		Short: "Start a stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireStackName(name); err != nil {
				return err
			}
			if !runs.Exists(name) {
				return fmt.Errorf("stack %q does not exist", name)
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()
			if err := compose.Up(ctx, name); err != nil {
				return err
			}
			_ = runs.Touch(name)
			fmt.Printf("Started stack %s\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "stack name")
	return cmd
}

func newStopCmd(runs *stack.RunStore, compose *docker.Compose) *cobra.Command {
	name := ""
	cmd := &cobra.Command{
		Use:   "stop --name <stack>",
		Short: "Stop a stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireStackName(name); err != nil {
				return err
			}
			if !runs.Exists(name) {
				return fmt.Errorf("stack %q does not exist", name)
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()
			if err := compose.Stop(ctx, name); err != nil {
				return err
			}
			_ = runs.Touch(name)
			fmt.Printf("Stopped stack %s\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "stack name")
	return cmd
}

func newRestartCmd(runs *stack.RunStore, compose *docker.Compose) *cobra.Command {
	name := ""
	cmd := &cobra.Command{
		Use:   "restart --name <stack>",
		Short: "Restart a stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireStackName(name); err != nil {
				return err
			}
			if !runs.Exists(name) {
				return fmt.Errorf("stack %q does not exist", name)
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()
			if err := compose.Restart(ctx, name); err != nil {
				return err
			}
			_ = runs.Touch(name)
			fmt.Printf("Restarted stack %s\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "stack name")
	return cmd
}

func newLogsCmd(runs *stack.RunStore, compose *docker.Compose) *cobra.Command {
	name := ""
	service := ""
	follow := false
	cmd := &cobra.Command{
		Use:   "logs --name <stack>",
		Short: "Show stack logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireStackName(name); err != nil {
				return err
			}
			if !runs.Exists(name) {
				return fmt.Errorf("stack %q does not exist", name)
			}
			out, err := compose.Logs(cmd.Context(), name, service, follow)
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "stack name")
	cmd.Flags().StringVar(&service, "service", "", "service name filter")
	cmd.Flags().BoolVar(&follow, "follow", false, "follow logs")
	return cmd
}

func newRemoveCmd(runs *stack.RunStore, compose *docker.Compose) *cobra.Command {
	name := ""
	yes := false
	all := false
	cmd := &cobra.Command{
		Use:   "remove --name <stack>",
		Short: "Remove a stack and delete its run directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if all {
				return removeAll(cmd.Context(), runs, compose, yes)
			}
			if err := requireStackName(name); err != nil {
				return err
			}
			if !runs.Exists(name) {
				return fmt.Errorf("stack %q does not exist", name)
			}
			if !yes {
				ok, err := tui.Confirm("Remove stack?", fmt.Sprintf("This will delete stack %q and its configuration. This is destructive.", name), false)
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("remove canceled")
				}
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()
			if err := compose.Down(ctx, name); err != nil {
				return err
			}
			if err := runs.Delete(name); err != nil {
				return err
			}
			fmt.Printf("Removed stack %s\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "stack name")
	cmd.Flags().BoolVar(&yes, "yes", false, "confirm removal")
	cmd.Flags().BoolVar(&all, "all", false, "remove all stacks")
	return cmd
}

func removeAll(ctx context.Context, runs *stack.RunStore, compose *docker.Compose, yes bool) error {
	metas, err := runs.List()
	if err != nil {
		return fmt.Errorf("list stacks: %w", err)
	}
	if len(metas) == 0 {
		fmt.Println("No stacks to remove")
		return nil
	}
	if !yes {
		ok, err := tui.Confirm("Remove all stacks?", fmt.Sprintf("This will delete all %d stacks and their configuration. This is destructive.", len(metas)), false)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("remove canceled")
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	for _, m := range metas {
		if err := compose.Down(ctx, m.Name); err != nil {
			fmt.Printf("Warning: failed to stop stack %s: %v\n", m.Name, err)
		}
		if err := runs.Delete(m.Name); err != nil {
			fmt.Printf("Warning: failed to delete stack %s: %v\n", m.Name, err)
			continue
		}
		fmt.Printf("Removed stack %s\n", m.Name)
	}
	return nil
}

package app

import (
	"fmt"

	"github.com/openhoo/vibecontainer/internal/docker"
	"github.com/openhoo/vibecontainer/internal/stack"
	"github.com/openhoo/vibecontainer/internal/tui"
	"github.com/spf13/cobra"
)

func newStatusCmd(runs *stack.RunStore, compose *docker.Compose) *cobra.Command {
	name := ""
	cmd := &cobra.Command{
		Use:   "status --name <stack>",
		Short: "Show service status for a stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireStackName(name); err != nil {
				return err
			}
			if !runs.Exists(name) {
				return fmt.Errorf("stack %q does not exist", name)
			}
			statuses, err := compose.Status(cmd.Context(), name)
			if err != nil {
				return err
			}
			if len(statuses) == 0 {
				fmt.Println("No services found")
				return nil
			}
			tui.RenderStatus(statuses)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "stack name")
	return cmd
}

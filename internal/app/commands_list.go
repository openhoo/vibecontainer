package app

import (
	"fmt"

	"github.com/openhoo/vibecontainer/internal/docker"
	"github.com/openhoo/vibecontainer/internal/stack"
	"github.com/openhoo/vibecontainer/internal/tui"
	"github.com/spf13/cobra"
)

const labelStack = "com.openhoo.vibecontainer.stack"

func newListCmd(runs *stack.RunStore, compose *docker.Compose) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List managed stacks",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			metas, err := runs.List()
			if err != nil {
				return err
			}
			managed, err := compose.ListManagedContainers(ctx)
			if err != nil {
				return err
			}
			stateByStack := map[string][]string{}
			for _, c := range managed {
				stackName := c.Labels[labelStack]
				if stackName == "" {
					continue
				}
				stateByStack[stackName] = append(stateByStack[stackName], c.State)
			}
			if len(metas) == 0 {
				fmt.Println("No managed stacks found")
				return nil
			}
			tui.RenderList(metas, stateByStack)
			return nil
		},
	}
}

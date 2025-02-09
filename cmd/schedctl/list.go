package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"schedctl/internal/containerd"
)

func NewListCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "list",
		Short: "list running schedulers",
		RunE:  list,
	}

	return startCmd
}

func list(cmd *cobra.Command, arguments []string) error {
	containersList, err := containerd.List()
	if err != nil {
		panic(err)
	}

	for _, container := range containersList {
		fmt.Printf("pid: %s, id: %s, name: %s", container.PID, container.ID, container.Name)
	}

	return nil
}

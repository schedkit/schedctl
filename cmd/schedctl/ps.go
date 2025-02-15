package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"schedctl/internal/containerd"
)

func NewPsCmd() *cobra.Command {
	psCmd := &cobra.Command{
		Use:   "ps",
		Short: "list running schedulers",
		RunE:  ps,
	}

	return psCmd
}

func ps(cmd *cobra.Command, arguments []string) error {
	containersList, err := containerd.List()
	if err != nil {
		panic(err)
	}

	for _, container := range containersList {
		fmt.Printf("pid: %d, id: %s, name: %s", container.PID, container.ID, container.Name)
	}

	return nil
}

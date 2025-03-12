package cmd

import (
	"github.com/spf13/cobra"

	"schedctl/internal/containerd"
	"schedctl/internal/output"
)

func NewPsCmd() *cobra.Command {
	psCmd := &cobra.Command{
		Use:   "ps",
		Short: "list running schedulers",
		RunE:  ps,
	}

	return psCmd
}

func ps(_ *cobra.Command, _ []string) error {
	client, err := containerd.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	containersList, err := containerd.List(client)
	if err != nil {
		panic(err)
	}

	for _, container := range containersList {
		_, _ = output.Out("pid: %d, id: %s, name: %s", container.PID, container.ID, container.Name)
	}

	return nil
}

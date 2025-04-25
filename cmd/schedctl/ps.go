package cmd

import (
	"github.com/spf13/cobra"

	"schedctl/internal/constants"
	"schedctl/internal/containerd"
	"schedctl/internal/containers"
	"schedctl/internal/output"
)

func NewPsCmd() *cobra.Command {
	psCmd := &cobra.Command{
		Use:   "ps",
		Short: "list running schedulers",
		RunE:  ps,
	}

	psCmd.PersistentFlags().StringP("driver", "d", "containerd", "The driver to use: containerd, podman")

	return psCmd
}

func ps(cmd *cobra.Command, _ []string) error {
	driver := cmd.Flags().Lookup("driver").Value.String()
	client, err := containerd.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	containersList := make([]containers.Container, 0)

	if driver == constants.CONTAINERD {
		containerdList, err := containerd.List(client)
		if err != nil {
			panic(err)
		}
		containersList = append(containersList, containerdList...)
	}

	if driver == constants.PODMAN {
	}

	for _, container := range containersList {
		_, _ = output.Out("pid: %d, id: %s, name: %s", container.PID, container.ID, container.Name)
	}

	return nil
}

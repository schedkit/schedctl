package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"schedctl/internal/containerd"
)

func NewStopCmd() *cobra.Command {
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "stop a scheduler",
		RunE:  stop,
	}

	stopCmd.PersistentFlags().StringP("driver", "d", "containerd", "The driver to use: containerd, podman")

	return stopCmd
}

func stop(cmd *cobra.Command, _ []string) error {
	id := cmd.Flags().Args()[0]
	driver := cmd.Flags().Lookup("driver").Value.String()

	if driver == "containerd" {
		client, err := containerd.NewClient()
		if err != nil {
			panic(err)
		}
		defer client.Close()

		err = containerd.Stop(client, id)
		if err != nil {
			return fmt.Errorf("failed to stop the container: %w", err)
		}
	}

	if driver == "podman" {
	}

	return nil
}

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

	return stopCmd
}

func stop(cmd *cobra.Command, arguments []string) error {
	id := cmd.Flags().Args()[0]

	err := containerd.Stop(id)
	if err != nil {
		fmt.Errorf("failed to stop the container: %v", err)
		return err
	}

	return nil
}

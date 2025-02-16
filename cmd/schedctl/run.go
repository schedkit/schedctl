package cmd

import (
	"schedctl/internal/containerd"
	"schedctl/internal/schedulers"

	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "run",
		Short: "Run a specific scheduler",
		RunE:  run,
	}

	return startCmd
}

func run(cmd *cobra.Command, arguments []string) error {
	schedulerId := cmd.Flags().Args()[0]

	image, err := schedulers.GetScheduler(schedulerId)
	if err != nil {
		return err
	}

	err = containerd.Run(image, schedulerId)
	if err != nil {
		return err
	}

	return nil
}

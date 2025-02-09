package cmd

import (
	"schedctl/internal/containerd"

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
	src := cmd.Flags().Args()[0]

	err := containerd.Run(src, "demo-container")
	if err != nil {
		return err
	}

	return nil
}

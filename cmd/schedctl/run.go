package cmd

import (
	"schedctl/internal/containerd"
	"schedctl/internal/schedulers"

	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	var Attach bool

	startCmd := &cobra.Command{
		Use:   "run",
		Short: "Run a specific scheduler",
		RunE: func(cmd *cobra.Command, arguments []string) error {
			return run(cmd, arguments, Attach)
		},
	}

	startCmd.Flags().BoolVarP(&Attach, "attach", "a", false, "attach to the current process instead of detaching")

	return startCmd
}

func run(cmd *cobra.Command, _ []string, attach bool) error {
	schedulerID := cmd.Flags().Args()[0]

	// connect to rootful containerd
	client, err := containerd.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	image, err := schedulers.GetScheduler(schedulerID)
	if err != nil {
		return err
	}

	err = containerd.Run(client, image, schedulerID, attach, true)
	if err != nil {
		return err
	}

	return nil
}

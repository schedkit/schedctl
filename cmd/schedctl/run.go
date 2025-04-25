package cmd

import (
	"schedctl/internal/containerd"
	"schedctl/internal/output"
	"schedctl/internal/podman"
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
	startCmd.PersistentFlags().StringP("driver", "d", "containerd", "The driver to use: containerd, podman")

	return startCmd
}

func run(cmd *cobra.Command, _ []string, attach bool) error {
	driver := cmd.Flags().Lookup("driver").Value.String()
	schedulerID := cmd.Flags().Args()[0]

	image, err := schedulers.GetScheduler(schedulerID)
	if err != nil {
		return err
	}

	if driver == "containerd" {
		// connect to rootful containerd
		client, err := containerd.NewClient()
		if err != nil {
			panic(err)
		}
		defer client.Close()

		err = containerd.Run(client, image, schedulerID, attach, true)
		if err != nil {
			return err
		}
	}

	if driver == "podman" {
		err := podman.Run(image)
		if err != nil {
			panic(err)
		}

		output.Out("Container %s started successfully\n", image)
	}

	return nil
}

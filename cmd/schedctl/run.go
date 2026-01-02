package cmd

import (
	"fmt"
	"schedctl/internal/constants"
	"schedctl/internal/containerd"
	"schedctl/internal/output"
	"schedctl/internal/podman"
	"schedctl/internal/schedulers"

	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	var Attach bool

	startCmd := &cobra.Command{
		Use:   "run SCHEDULER [-- ARGS...]",
		Short: "Run a specific scheduler",
		Long: `Run a specific scheduler with optional arguments.

Arguments after -- are passed to the scheduler container.

Examples:
  schedctl run scx_rusty
  schedctl run scx_rusty -- --verbose
  schedctl run --attach scx_rusty -- --mode=performance --interval=100`,
		Args: func(cmd *cobra.Command, args []string) error {
			dashIndex := cmd.ArgsLenAtDash()
			if dashIndex == -1 {
				if len(args) != 1 {
					return fmt.Errorf("exactly one scheduler ID required")
				}
			} else {
				if dashIndex != 1 {
					return fmt.Errorf("exactly one scheduler ID required before --")
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, arguments []string) error {
			return run(cmd, arguments, Attach)
		},
	}

	startCmd.Flags().BoolVarP(&Attach, "attach", "a", false, "attach to the current process instead of detaching")
	startCmd.PersistentFlags().StringP("driver", "d", "podman", "The driver to use: containerd, podman")

	return startCmd
}

func run(cmd *cobra.Command, args []string, attach bool) error {
	driver := cmd.Flags().Lookup("driver").Value.String()
	schedulerID := args[0]

	var containerArgs []string
	if cmd.ArgsLenAtDash() >= 0 {
		containerArgs = args[cmd.ArgsLenAtDash():]
	}

	result, err := schedulers.GetScheduler(schedulerID)
	if err != nil {
		return err
	}

	switch result.Source {
	case schedulers.SourceManifest:
		_, _ = output.Out("Running scheduler '%s' from manifest: %s\n", schedulerID, result.ImageURI)
	case schedulers.SourceDirect:
		_, _ = output.Out("Running container image: %s\n", result.ImageURI)
	}

	if len(containerArgs) > 0 {
		_, _ = output.Out("With arguments: %v\n", containerArgs)
	}

	if driver == constants.CONTAINERD {
		// connect to rootful containerd
		client, err := containerd.NewClient()
		if err != nil {
			panic(err)
		}
		defer client.Close()

		err = containerd.Run(client, result.ImageURI, schedulerID, attach, true, containerArgs)
		if err != nil {
			return err
		}
	}

	if driver == constants.PODMAN {
		err := podman.Run(result.ImageURI, schedulerID, containerArgs)
		if err != nil {
			panic(err)
		}

		_, _ = output.Out("Container %s started successfully\n", result.ImageURI)
	}

	return nil
}

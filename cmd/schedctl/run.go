package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"schedctl/internal/constants"
	"schedctl/internal/containerd"
	"schedctl/internal/output"
	"schedctl/internal/podman"
	"schedctl/internal/schedulers"
)

func NewRunCmd() *cli.Command {
	return &cli.Command{
		Name:      "run",
		Usage:     "Run a specific scheduler",
		ArgsUsage: "SCHEDULER [-- ARGS...]",
		Description: `Run a specific scheduler with optional arguments.

Arguments after -- are passed to the scheduler container.

Examples:
  schedctl run scx_rusty
  schedctl run scx_rusty -- --verbose
  schedctl run --attach scx_rusty -- --mode=performance --interval=100`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "attach",
				Aliases:  []string{"a"},
				Usage:    "attach to the current process instead of detaching",
				Local:    true,
				Category: categoryProcess,
			},
		},
		Action: runAction,
	}
}

func runAction(_ context.Context, cmd *cli.Command) error {
	args := cmd.Args().Slice()
	if len(args) == 0 {
		return fmt.Errorf("exactly one scheduler ID required")
	}

	schedulerID := args[0]
	containerArgs := args[1:]
	driver := cmd.String("driver")
	attach := cmd.Bool("attach")

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
		err := podman.Run(result.ImageURI, schedulerID, attach, containerArgs)
		if err != nil {
			panic(err)
		}

		if !attach {
			_, _ = output.Out("Container %s started successfully\n", result.ImageURI)
		}
	}

	return nil
}

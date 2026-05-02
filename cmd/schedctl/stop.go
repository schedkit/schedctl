package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"schedctl/internal/constants"
	"schedctl/internal/containerd"
	"schedctl/internal/podman"
)

func NewStopCmd() *cli.Command {
	return &cli.Command{
		Name:      "stop",
		Usage:     "stop a scheduler",
		ArgsUsage: "ID",
		Action:    stopAction,
	}
}

func stopAction(_ context.Context, cmd *cli.Command) error {
	args := cmd.Args().Slice()
	if len(args) == 0 {
		return fmt.Errorf("container ID required")
	}

	id := args[0]
	driver := cmd.String("driver")

	if driver == constants.CONTAINERD {
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

	if driver == constants.PODMAN {
		err := podman.Stop(id)
		if err != nil {
			return fmt.Errorf("failed to stop the container: %w", err)
		}
	}

	return nil
}

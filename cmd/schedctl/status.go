package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"schedctl/internal/constants"
	"schedctl/internal/containerd"
	"schedctl/internal/containers"
	"schedctl/internal/podman"
	"schedctl/internal/sched_ext"
	"schedctl/internal/status"
)

const categoryDisplay = "Output:"

func NewStatusCmd() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "report what schedctl is managing and the kernel sched_ext state",
		Description: `Report the scheduler currently managed by schedctl, including the
container image and digest it was started from, the start time, the
container runtime in use, and the attachment state reported by the kernel
via /sys/kernel/sched_ext/.

Exit codes:
  0  scheduler running normally, or no scheduler running
  1  generic error (failed to query the runtime, etc.)
  2  schedctl's view and the kernel's view do not agree

See schedctl-status(1) for the full JSON schema.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Usage:    "output format: text, json",
				Value:    "text",
				Local:    true,
				Category: categoryDisplay,
			},
		},
		Action: statusAction,
	}
}

func statusAction(_ context.Context, cmd *cli.Command) error {
	driver := cmd.String("driver")
	format := cmd.String("output")

	if format != "text" && format != "json" {
		return fmt.Errorf("unsupported --output value %q (expected: text, json)", format)
	}

	managed, err := listManaged(driver)
	if err != nil {
		return fmt.Errorf("failed to query %s: %w", driver, err)
	}

	kernel, err := sched_ext.Read()
	if err != nil {
		return fmt.Errorf("failed to read kernel sched_ext state: %w", err)
	}

	report := status.Build(driver, managed, kernel)

	switch format {
	case "json":
		if err := status.WriteJSON(os.Stdout, report); err != nil {
			return err
		}
	default:
		status.WriteText(os.Stdout, report)
	}

	if report.IsDiscrepancy() {
		return cli.Exit("", 2)
	}
	return nil
}

func listManaged(driver string) ([]containers.Container, error) {
	switch driver {
	case constants.PODMAN:
		return podman.List()
	case constants.CONTAINERD:
		client, err := containerd.NewClient()
		if err != nil {
			return nil, err
		}
		defer client.Close()
		return containerd.List(client)
	default:
		return nil, fmt.Errorf("unknown driver %q", driver)
	}
}

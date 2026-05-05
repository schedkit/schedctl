package cmd

import (
	"context"
	"fmt"
	"sort"

	"github.com/urfave/cli/v3"

	"schedctl/internal/output"
	"schedctl/internal/schedulers"
)

func NewVersionsCmd() *cli.Command {
	return &cli.Command{
		Name:      "versions",
		Usage:     "list available versions for a scheduler",
		ArgsUsage: "SCHEDULER",
		Description: `Query the container registry for all available versions (tags)
of the given scheduler.

Examples:
  schedctl versions scx_rusty
  schedctl versions ghcr.io/myorg/custom-scheduler`,
		Action: versionsAction,
	}
}

func versionsAction(_ context.Context, cmd *cli.Command) error {
	args := cmd.Args().Slice()
	if len(args) == 0 {
		return fmt.Errorf("exactly one scheduler ID required")
	}

	schedulerID := args[0]

	versions, err := schedulers.ListVersions(schedulerID)
	if err != nil {
		return err
	}

	sort.Strings(versions)
	for _, v := range versions {
		_, _ = output.Out("%s\n", v)
	}

	return nil
}

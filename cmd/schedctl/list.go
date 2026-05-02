package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"schedctl/internal/output"
	"schedctl/internal/schedulers"
)

func NewListCmd() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "list available schedulers",
		Action: listAction,
	}
}

func listAction(_ context.Context, _ *cli.Command) error {
	for key := range schedulers.List() {
		_, _ = output.Out("%s\n", key)
	}

	return nil
}

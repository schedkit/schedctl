package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"schedctl/internal/info"
	"schedctl/internal/schedulers"
)

func NewInfoCmd() *cli.Command {
	return &cli.Command{
		Name:      "info",
		Usage:     "show metadata for a scheduler image",
		ArgsUsage: "SCHEDULER",
		Description: `Show metadata embedded in a scheduler image: title, description,
documentation link, source repository, minimum kernel version, and the
other standard OCI annotations the image exposes.

SCHEDULER is either a name from schedctl list or a fully-qualified
container image reference. The image is queried from the registry; no
container runtime is required.

Examples:
  schedctl info scx_rusty
  schedctl info scx_rusty --version v1.0.0
  schedctl info ghcr.io/myorg/custom-scheduler -o json`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "version",
				Usage:    "image tag to inspect, e.g. v1.0.0",
				Local:    true,
				Category: categoryProcess,
			},
			&cli.StringFlag{
				Name:     flagOutput,
				Aliases:  []string{"o"},
				Usage:    flagOutputUsage,
				Value:    outputText,
				Local:    true,
				Category: categoryDisplay,
			},
		},
		Action: infoAction,
	}
}

func infoAction(_ context.Context, cmd *cli.Command) error {
	args := cmd.Args().Slice()
	if len(args) == 0 {
		return fmt.Errorf("exactly one scheduler ID required")
	}

	schedulerID := args[0]
	version := cmd.String("version")
	format := cmd.String(flagOutput)

	if format != outputText && format != outputJSON {
		return fmt.Errorf("unsupported --output value %q (expected: %s, %s)", format, outputText, outputJSON)
	}

	result, err := schedulers.GetScheduler(schedulerID, version)
	if err != nil {
		return err
	}

	raw, err := schedulers.Inspect(result.ImageURI)
	if err != nil {
		return err
	}

	report := info.Build(schedulerID, result, raw)

	switch format {
	case outputJSON:
		return info.WriteJSON(os.Stdout, report)
	default:
		info.WriteText(os.Stdout, report)
		return nil
	}
}

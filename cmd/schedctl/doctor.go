package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"schedctl/internal/doctor"
)

const categoryDiagnostics = "Diagnostics:"

func NewDoctorCmd() *cli.Command {
	return &cli.Command{
		Name:  "doctor",
		Usage: "check host readiness for sched_ext schedulers",
		Description: "Run sanity checks against the host (kernel, capabilities, container runtime)." +
			" Exits non-zero when any blocking check fails.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Usage:    "output format: text, json",
				Value:    outputText,
				Local:    true,
				Category: categoryDiagnostics,
			},
		},
		Action: doctorAction,
	}
}

func doctorAction(_ context.Context, cmd *cli.Command) error {
	report := doctor.Run(doctor.DefaultChecks())

	switch cmd.String("output") {
	case outputJSON:
		if err := doctor.WriteJSON(os.Stdout, report); err != nil {
			return err
		}
	case outputText, "":
		if err := doctor.WriteText(os.Stdout, report); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported output format %q (expected %s or %s)",
			cmd.String("output"), outputText, outputJSON)
	}

	if report.HasBlockingFailures() {
		return cli.Exit("", 1)
	}
	return nil
}

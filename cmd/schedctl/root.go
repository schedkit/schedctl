package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

const (
	categoryRuntime = "Container runtime:"
	categoryProcess = "Process control:"
)

const (
	outputText = "text"
	outputJSON = "json"
)

const (
	flagOutput      = "output"
	flagOutputUsage = "output format: text, json"
)

func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func NewRootCmd() *cli.Command {
	return &cli.Command{
		Name:                  "schedctl",
		Usage:                 "Plug and play bpf schedulers for fun and profit",
		Description:           "Plug and play bpf schedulers for fun and profit",
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "driver",
				Aliases:  []string{"d"},
				Usage:    "container runtime to use: containerd, podman",
				Value:    "podman",
				Category: categoryRuntime,
			},
		},
		Commands: []*cli.Command{
			NewListCmd(),
			NewRunCmd(),
			NewPsCmd(),
			NewStopCmd(),
			NewStatusCmd(),
			NewDoctorCmd(),
			NewVersionsCmd(),
			NewInfoCmd(),
		},
	}
}

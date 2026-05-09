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
	"schedctl/internal/verify"
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
  schedctl run scx_rusty --version v1.0.0
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
			&cli.StringFlag{
				Name:     "version",
				Usage:    "scheduler version (image tag) to run, e.g. v1.0.0",
				Local:    true,
				Category: categoryProcess,
			},
			&cli.StringFlag{
				Name:     "trust-policy",
				Usage:    "path to a YAML trust policy for image signature verification",
				Sources:  cli.EnvVars("SCHEDCTL_TRUST_POLICY"),
				Local:    true,
				Category: categoryProcess,
			},
			&cli.BoolFlag{
				Name:     "allow-unsigned",
				Usage:    "skip signature verification (NOT recommended; images run with elevated caps and load eBPF)",
				Sources:  cli.EnvVars("SCHEDCTL_ALLOW_UNSIGNED"),
				Local:    true,
				Category: categoryProcess,
			},
		},
		Action: runAction,
	}
}

func runAction(ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args().Slice()
	if len(args) == 0 {
		return fmt.Errorf("exactly one scheduler ID required")
	}

	schedulerID := args[0]
	containerArgs := args[1:]
	driver := cmd.String("driver")
	attach := cmd.Bool("attach")
	version := cmd.String("version")
	trustPolicyPath := cmd.String("trust-policy")
	allowUnsigned := cmd.Bool("allow-unsigned")

	result, err := schedulers.GetScheduler(schedulerID, version)
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

	imageRef := result.ImageURI
	if allowUnsigned {
		_, _ = output.Out("WARNING: --allow-unsigned set; skipping signature verification for %s\n", result.ImageURI)
	} else {
		policy, err := verify.LoadPolicy(trustPolicyPath)
		if err != nil {
			return fmt.Errorf("trust policy: %w", err)
		}
		verified, err := verify.Image(ctx, result.ImageURI, policy)
		if err != nil {
			return fmt.Errorf(
				"image signature verification failed for %s: %w (pass --allow-unsigned to override)",
				result.ImageURI, err,
			)
		}
		signer := "unknown"
		if len(verified.Signers) > 0 {
			signer = verified.Signers[0].Describe()
		}
		_, _ = output.Out("Verified %s (signer: %s)\n", verified.ImageRef, signer)
		imageRef = verified.ImageRef
	}

	if driver == constants.CONTAINERD {
		client, err := containerd.NewClient()
		if err != nil {
			panic(err)
		}
		defer client.Close()

		err = containerd.Run(client, imageRef, schedulerID, attach, true, containerArgs)
		if err != nil {
			return err
		}
	}

	if driver == constants.PODMAN {
		err := podman.Run(imageRef, schedulerID, attach, containerArgs)
		if err != nil {
			panic(err)
		}

		if !attach {
			_, _ = output.Out("Container %s started successfully\n", imageRef)
		}
	}

	return nil
}

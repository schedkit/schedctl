package cmd

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"

	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "run",
		Short: "Run a specific scheduler",
		RunE:  run,
	}

	return startCmd
}

func run(cmd *cobra.Command, arguments []string) error {
	cleanup()

	src := cmd.Flags().Args()[0]
	// Create a new context with namespace
	ctx := namespaces.WithNamespace(context.Background(), "schedulers")

	// Create a new containerd client
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Get the image reference

	// Pull the image
	img, err := client.Pull(ctx, src, containerd.WithPullUnpack)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	// Create a new container
	container, err := client.NewContainer(
		ctx,
		"demo-container",
		containerd.WithNewSnapshot("demo-snapshot", img),
		containerd.WithNewSpec(oci.WithImageConfig(img), oci.WithPrivileged),
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)

	// Create a task
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	defer task.Delete(ctx)

	// Start the task
	if err := task.Start(ctx); err != nil {
		return fmt.Errorf("failed to start task: %w", err)
	}

	fmt.Println("Task started, PID:", task.Pid())

	// Wait for the task to exit
	exitStatusC, err := task.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for task: %w", err)
	}

	// Get the exit status
	status := <-exitStatusC
	code, _, err := status.Result()
	if err != nil {
		return fmt.Errorf("failed to get exit status: %w", err)
	}

	if code != 0 {
		return fmt.Errorf("container exited with status: %d", code)
	}

	return nil
}

func cleanup() error {
	// Create a new context with namespace
	ctx := namespaces.WithNamespace(context.Background(), "schedulers")

	// Create a new containerd client
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Get the snapshotter
	snapshotter := client.SnapshotService("overlayfs")

	// Remove the snapshot
	if err := snapshotter.Remove(ctx, "demo-snapshot"); err != nil {
		return fmt.Errorf("failed to remove snapshot: %w", err)
	}

	fmt.Println("Successfully removed demo-snapshot")

	container, err := client.LoadContainer(ctx, "demo-container")
	if err != nil {
		fmt.Errorf("failed to load container: %w", err)
	}

	if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		fmt.Errorf("failed to delete container: %w", err)
	}

	return nil
}

func BoolPointer(b bool) *bool {
	return &b
}

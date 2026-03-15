package containerd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/containerd/containerd/v2/pkg/oci"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"schedctl/internal/containers"
	"schedctl/internal/output"
)

func generateRandomSuffix() (string, error) {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func NewClient() (*containerd.Client, error) {
	// TODO make this configurable if needed
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client, nil
}

func List(client *containerd.Client) ([]containers.Container, error) {
	// Create a new context with namespace
	ctx := namespaces.WithNamespace(context.Background(), "schedkit")

	listedContainers := []containers.Container{}

	// List all containers in the specified namespace
	containerdContainers, err := client.Containers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	// Print container details
	for _, container := range containerdContainers {
		id := container.ID()

		task, err := container.Task(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get container task: %w", err)
		}

		pid := int(task.Pid())

		listedContainer := containers.Container{
			PID: pid,
			ID:  id,
		}

		listedContainers = append(listedContainers, listedContainer)
	}

	return listedContainers, nil
}

func Stop(client *containerd.Client, containerID string) error {
	ctx := namespaces.WithNamespace(context.Background(), "schedkit")

	container, err := client.LoadContainer(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to load container: %w", err)
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	_ = task.Kill(ctx, 9) // SIGKILL all the things
	exitChan, err := task.Wait(ctx)
	if err != nil {
		_, _ = output.Out("Failed waiting for the task to exit")
	}
	<-exitChan

	_, err = task.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	err = container.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	_, _ = output.Out("Scheduler %s stopped successfully \n", containerID)

	return nil
}

func Run(client *containerd.Client, image, id string, attach bool, privileged bool, args []string) error {
	ctx := namespaces.WithNamespace(context.Background(), "schedkit")

	img, err := client.Pull(ctx, image, containerd.WithPullUnpack)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	var specOpts []oci.SpecOpts
	specOpts = append(specOpts, oci.WithImageConfig(img))

	if privileged {
		specOpts = append(specOpts, oci.WithPrivileged)
	}

	if len(args) > 0 {
		specOpts = append(specOpts, oci.WithProcessArgs(args...))
	}

	// Ensure /var/run/scx exists on host for scx scheduler stats socket sharing
	if err := os.MkdirAll("/var/run/scx", 0755); err != nil {
		return fmt.Errorf("failed to create /var/run/scx: %w", err)
	}

	// Add bind mount for scx stats socket directory
	specOpts = append(specOpts, oci.WithMounts([]specs.Mount{
		{
			Source:      "/var/run/scx",
			Destination: "/var/run/scx",
			Type:        "bind",
			Options:     []string{"rbind", "rw"},
		},
	}))

	specOption := containerd.WithNewSpec(specOpts...)

	containerID := id
	snapshotID := fmt.Sprintf("%s-snapshot\n", id)

	container, err := client.NewContainer(
		ctx,
		containerID,
		containerd.WithNewSnapshot(snapshotID, img),
		specOption,
	)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			suffix, suffixErr := generateRandomSuffix()
			if suffixErr != nil {
				return fmt.Errorf("failed to generate random suffix: %w", suffixErr)
			}
			containerID = fmt.Sprintf("%s-%s", id, suffix)
			snapshotID = fmt.Sprintf("%s-snapshot\n", containerID)
			container, err = client.NewContainer(
				ctx,
				containerID,
				containerd.WithNewSnapshot(snapshotID, img),
				specOption,
			)
			if err != nil {
				return fmt.Errorf("failed to create container with random name: %w", err)
			}
		} else {
			return fmt.Errorf("failed to create container: %w", err)
		}
	}
	defer func() { _ = container.Delete(ctx, containerd.WithSnapshotCleanup) }()

	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	defer func() { _, _ = task.Delete(ctx) }()

	err = task.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start task: %w", err)
	}

	_, _ = output.Out("Task started, PID: %d\n", task.Pid())

	if attach {
		exitStatusC, err := task.Wait(ctx)
		if err != nil {
			return fmt.Errorf("failed to wait for task: %w", err)
		}

		status := <-exitStatusC
		code, _, err := status.Result()
		if err != nil {
			return fmt.Errorf("failed to get exit status: %w", err)
		}

		if code != 0 {
			return fmt.Errorf("container exited with status: %d", code)
		}
	}

	return nil
}

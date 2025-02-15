package containerd

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"

	"schedctl/internal/containers"
)

func List() ([]containers.Container, error) {
	// Create a new context with namespace
	ctx := namespaces.WithNamespace(context.Background(), "schedkit")

	listedContainers := []containers.Container{}

	// Create a new containerd client
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

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

		pid := task.Pid()

		listedContainer := containers.Container{
			PID: pid,
			ID:  id,
		}

		listedContainers = append(listedContainers, listedContainer)
	}

	return listedContainers, nil
}

func Stop(containerId string) error {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := namespaces.WithNamespace(context.Background(), "schedkit")

	container, err := client.LoadContainer(ctx, containerId)
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
		fmt.Printf("Failed waiting for the task to exit")
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

	fmt.Printf("Scheduler %s stopped successfully \n", containerId)

	return nil
}

func Run(image, id string) error {
	// Create a new context with namespace
	ctx := namespaces.WithNamespace(context.Background(), "schedkit")

	// Create a new containerd client
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Pull the image
	img, err := client.Pull(ctx, image, containerd.WithPullUnpack)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	// Create a new container
	container, err := client.NewContainer(
		ctx,
		id,
		containerd.WithNewSnapshot(fmt.Sprintf("%s-snapshot\n", id), img),
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
	err = task.Start(ctx)
	if err != nil {
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

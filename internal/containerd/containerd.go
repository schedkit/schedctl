package containerd

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"

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

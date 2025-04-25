package podman

import (
	"context"
	"fmt"
	"strings"

	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/specgen"
)

func Run(image string) error {
	ctx := context.Background()
	privileged := true

	client, err := bindings.NewConnection(ctx, "unix:/run/podman/podman.sock")
	if err != nil {
		return fmt.Errorf("failed to create Podman connection: %w", err)
	}

	_, err = images.Pull(client, image, nil)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	spec := specgen.NewSpecGenerator(image, false)
	spec.Name = beforeColon(image)
	spec.CgroupNS.Value = "schedkit"
	spec.Privileged = &privileged

	createResponse, err := containers.CreateWithSpec(client, spec, nil)
	if err != nil {
		return fmt.Errorf("failed to create container spec: %w", err)
	}

	if err := containers.Start(client, createResponse.ID, nil); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	return nil
}

func Stop(container string) error {
	// Create a new context
	ctx := context.Background()

	// Create a new Podman connection
	conn, err := bindings.NewConnection(ctx, "unix:/run/podman/podman.sock")
	if err != nil {
		return fmt.Errorf("failed to create Podman connection: %w", err)
	}

	err = containers.Stop(conn, container, nil)
	if err != nil {
		return fmt.Errorf("failed to stop container %s: %w", container, err)
	}

	fmt.Printf("Container %s stopped successfully\n", container)
	return nil
}

func beforeColon(input string) string {
	parts := strings.SplitN(input, ":", 2)
	return parts[0]
}

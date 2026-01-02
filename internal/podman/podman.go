package podman

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/containers/podman/v5/pkg/bindings"
	podman_containers "github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/specgen"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"schedctl/internal/containers"
)

func generateRandomSuffix() (string, error) {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func Run(image, id string, args []string) error {
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
	spec.Name = id
	spec.Privileged = &privileged
	spec.Labels = map[string]string{"provider": "schedkit"}
	spec.PidNS = specgen.Namespace{NSMode: specgen.NamespaceMode("host")}

	// Ensure /var/run/scx exists on host for scx scheduler stats socket sharing
	if err := os.MkdirAll("/var/run/scx", 0755); err != nil {
		return fmt.Errorf("failed to create /var/run/scx: %w", err)
	}

	// Add bind mount for scx stats socket directory
	spec.Mounts = []specs.Mount{
		{
			Source:      "/var/run/scx",
			Destination: "/var/run/scx",
			Type:        "bind",
			Options:     []string{"rbind", "rw"},
		},
	}

	if len(args) > 0 {
		spec.Command = args
	}

	createResponse, err := podman_containers.CreateWithSpec(client, spec, nil)
	if err != nil {
		if strings.Contains(err.Error(), "name is already in use") || strings.Contains(err.Error(), "already exists") {
			suffix, suffixErr := generateRandomSuffix()
			if suffixErr != nil {
				return fmt.Errorf("failed to generate random suffix: %w", suffixErr)
			}
			spec.Name = fmt.Sprintf("%s-%s", id, suffix)
			createResponse, err = podman_containers.CreateWithSpec(client, spec, nil)
			if err != nil {
				return fmt.Errorf("failed to create container with random name: %w", err)
			}
		} else {
			return fmt.Errorf("failed to create container spec: %w", err)
		}
	}

	if err := podman_containers.Start(client, createResponse.ID, nil); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	return nil
}

func Stop(container string) error {
	ctx := context.Background()

	// Create a new Podman connection
	conn, err := bindings.NewConnection(ctx, "unix:/run/podman/podman.sock")
	if err != nil {
		return fmt.Errorf("failed to create Podman connection: %w", err)
	}

	err = podman_containers.Stop(conn, container, nil)
	if err != nil {
		return fmt.Errorf("failed to stop container %s: %w", container, err)
	}

	_, err = podman_containers.Remove(conn, container, nil)
	if err != nil {
		return fmt.Errorf("failed to remove container %s: %w", container, err)
	}

	return nil
}

func List() ([]containers.Container, error) {
	ctx := context.Background()
	enabled := true

	listedContainers := []containers.Container{}

	// Create a new Podman connection
	conn, err := bindings.NewConnection(ctx, "unix:/run/podman/podman.sock")
	if err != nil {
		return nil, fmt.Errorf("failed to create Podman connection: %w", err)
	}

	options := podman_containers.ListOptions{
		All:     &enabled, // Only show running containers
		Filters: map[string][]string{"label": {"provider=schedkit"}},
	}

	podmanRunningContainers, err := podman_containers.List(conn, &options)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	for _, container := range podmanRunningContainers {
		ID := container.ID
		PID := container.Pid
		Name := container.Names[0]

		listedContainer := containers.Container{
			ID:   ID,
			PID:  PID,
			Name: Name,
		}

		listedContainers = append(listedContainers, listedContainer)
	}

	return listedContainers, nil
}

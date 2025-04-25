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
		fmt.Println(err)
		return err
	}
	s := specgen.NewSpecGenerator(image, false)
	s.Name = beforeColon(image)
	s.CgroupNS.Value = "schedkit"
	s.Privileged = &privileged

	createResponse, err := containers.CreateWithSpec(client, s, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if err := containers.Start(client, createResponse.ID, nil); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func beforeColon(input string) string {
	parts := strings.SplitN(input, ":", 2)
	return parts[0]
}

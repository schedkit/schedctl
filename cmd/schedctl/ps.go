package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli/v3"

	"schedctl/internal/constants"
	"schedctl/internal/containerd"
	"schedctl/internal/containers"
	"schedctl/internal/podman"
)

func NewPsCmd() *cli.Command {
	return &cli.Command{
		Name:   "ps",
		Usage:  "list running schedulers",
		Action: psAction,
	}
}

func psAction(_ context.Context, cmd *cli.Command) error {
	driver := cmd.String("driver")

	containersList := make([]containers.Container, 0)

	if driver == constants.CONTAINERD {
		client, err := containerd.NewClient()
		if err != nil {
			panic(err)
		}
		defer client.Close()

		containerdList, err := containerd.List(client)
		if err != nil {
			panic(err)
		}
		containersList = append(containersList, containerdList...)
	}

	if driver == constants.PODMAN {
		podmanList, err := podman.List()
		if err != nil {
			panic(err)
		}
		containersList = append(containersList, podmanList...)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PID\tID\tNAME")

	for _, container := range containersList {
		fmt.Fprintf(w, "%d\t%s\t%s\n", container.PID, container.ID, container.Name)
	}

	w.Flush()

	return nil
}

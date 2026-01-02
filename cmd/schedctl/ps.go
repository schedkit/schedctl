package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"schedctl/internal/constants"
	"schedctl/internal/containerd"
	"schedctl/internal/containers"
	"schedctl/internal/podman"
)

func NewPsCmd() *cobra.Command {
	psCmd := &cobra.Command{
		Use:   "ps",
		Short: "list running schedulers",
		RunE:  ps,
	}

	psCmd.PersistentFlags().StringP("driver", "d", "podman", "The driver to use: containerd, podman")

	return psCmd
}

func ps(cmd *cobra.Command, _ []string) error {
	driver := cmd.Flags().Lookup("driver").Value.String()

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

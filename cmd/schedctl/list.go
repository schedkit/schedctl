package cmd

import (
	"github.com/spf13/cobra"

	"schedctl/internal/output"
	"schedctl/internal/schedulers"
)

func NewListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list available schedulers",
		RunE:  list,
	}

	return listCmd
}

func list(_ *cobra.Command, _ []string) error {
	for key := range schedulers.List() {
		_, _ = output.Out("%s\n", key)
	}

	return nil
}

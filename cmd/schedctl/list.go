package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

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

func list(cmd *cobra.Command, arguments []string) error {
	for key, _ := range schedulers.List() {
		fmt.Printf("%s\n", key)
	}

	return nil
}

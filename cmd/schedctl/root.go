package cmd

import (
	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd := NewRootCmd()
	cobra.CheckErr(rootCmd.Execute())
}

func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "schedctl",
		Short: "Plug and play bpf schedulers for fun and profit",
		Long:  `Plug and play bpf schedulers for fun and profit`,
		Run:   func(_ *cobra.Command, _ []string) {},
	}

	rootCmd.AddCommand(NewRunCmd())
	rootCmd.AddCommand(NewPsCmd())
	rootCmd.AddCommand(NewStopCmd())
	rootCmd.AddCommand(NewListCmd())
	// TODO status

	return rootCmd
}

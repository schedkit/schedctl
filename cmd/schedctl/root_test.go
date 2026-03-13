package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"schedctl/cmd/schedctl"
)

func TestNewRootCmd(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	assert.NotNil(t, rootCmd)
	assert.Equal(t, "schedctl", rootCmd.Use)
	assert.Equal(t, "Plug and play bpf schedulers for fun and profit", rootCmd.Short)
	assert.Equal(t, "Plug and play bpf schedulers for fun and profit", rootCmd.Long)
}

func TestRootCmdHasSubcommands(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	expectedCommands := []string{"run", "ps", "stop", "list"}
	actualCommands := make([]string, 0, len(rootCmd.Commands()))

	for _, command := range rootCmd.Commands() {
		actualCommands = append(actualCommands, command.Name())
	}

	for _, expected := range expectedCommands {
		assert.Contains(t, actualCommands, expected, "Root command should have %s subcommand", expected)
	}
}

func TestRootCmdSubcommandCount(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	assert.Equal(t, 4, len(rootCmd.Commands()), "Root command should have exactly 4 subcommands")
}

func TestRootCmdExecuteWithoutArgs(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	err := rootCmd.Execute()
	assert.NoError(t, err)
}

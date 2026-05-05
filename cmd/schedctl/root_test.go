package cmd_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"

	cmd "schedctl/cmd/schedctl"
)

func lookupFlag(flags []cli.Flag, name string) cli.Flag {
	for _, f := range flags {
		for _, n := range f.Names() {
			if n == name {
				return f
			}
		}
	}
	return nil
}

func findSubcommand(c *cli.Command, name string) *cli.Command {
	for _, sub := range c.Commands {
		if sub.Name == name {
			return sub
		}
	}
	return nil
}

func TestNewRootCmd(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	assert.NotNil(t, rootCmd)
	assert.Equal(t, "schedctl", rootCmd.Name)
	assert.Equal(t, "Plug and play bpf schedulers for fun and profit", rootCmd.Usage)
	assert.Equal(t, "Plug and play bpf schedulers for fun and profit", rootCmd.Description)
}

func TestRootCmdHasSubcommands(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	expected := []string{"run", "ps", "stop", "list", "doctor", "status", "versions"}
	for _, name := range expected {
		assert.NotNil(t, findSubcommand(rootCmd, name), "Root command should have %s subcommand", name)
	}
}

func TestRootCmdSubcommandCount(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	assert.Equal(t, 7, len(rootCmd.Commands), "Root command should declare exactly 7 subcommands")
}

func TestRootCmdHasDriverPersistentFlag(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	driverFlag := lookupFlag(rootCmd.Flags, "driver")
	assert.NotNil(t, driverFlag, "root command should have 'driver' flag")

	stringFlag, ok := driverFlag.(*cli.StringFlag)
	assert.True(t, ok, "driver flag should be a StringFlag")
	assert.Equal(t, "podman", stringFlag.Value, "driver flag should default to 'podman'")
	assert.Contains(t, stringFlag.Aliases, "d", "driver flag should expose '-d' shorthand")
	assert.False(t, stringFlag.Local, "driver flag should be persistent (Local=false)")
	assert.Contains(t, stringFlag.Usage, "containerd")
	assert.Contains(t, stringFlag.Usage, "podman")
}

func TestRootCmdShellCompletionEnabled(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	assert.True(t, rootCmd.EnableShellCompletion, "shell completion should be enabled on the root command")
}

func TestRootCmdDriverFlagInheritedBySubcommands(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	err := rootCmd.Run(context.Background(), []string{"schedctl", "--driver=containerd", "list"})
	assert.NoError(t, err)
	assert.Equal(t, "containerd", rootCmd.String("driver"))
}

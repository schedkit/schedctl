package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cmd "schedctl/cmd/schedctl"
)

func TestNewInfoCmd(t *testing.T) {
	c := cmd.NewInfoCmd()

	assert.NotNil(t, c)
	assert.Equal(t, "info", c.Name)
	assert.Equal(t, "SCHEDULER", c.ArgsUsage)
	assert.NotEmpty(t, c.Usage)
	assert.NotEmpty(t, c.Description)
}

func TestInfoCmdHasOutputFlag(t *testing.T) {
	c := cmd.NewInfoCmd()

	outputFlag := lookupFlag(c.Flags, "output")
	assert.NotNil(t, outputFlag, "info command should expose --output")
	assert.Contains(t, outputFlag.Names(), "o", "info --output should expose '-o' shorthand")
}

func TestInfoCmdHasVersionFlag(t *testing.T) {
	c := cmd.NewInfoCmd()

	versionFlag := lookupFlag(c.Flags, "version")
	assert.NotNil(t, versionFlag, "info command should expose --version")
}

func TestInfoCmdRegisteredOnRoot(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	assert.NotNil(t, findSubcommand(rootCmd, "info"), "info subcommand must be registered on the root command")
}

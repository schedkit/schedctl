package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/urfave/cli/v3"

	cmd "schedctl/cmd/schedctl"
)

func TestNewRunCmd(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	assert.NotNil(t, runCmd)
	assert.Equal(t, "run", runCmd.Name)
	assert.Equal(t, "Run a specific scheduler", runCmd.Usage)
	assert.Equal(t, "SCHEDULER [-- ARGS...]", runCmd.ArgsUsage)
}

func TestRunCmdHasAttachFlag(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	attachFlag := lookupFlag(runCmd.Flags, "attach")
	assert.NotNil(t, attachFlag, "run command should have 'attach' flag")

	boolFlag, ok := attachFlag.(*cli.BoolFlag)
	assert.True(t, ok, "attach flag should be a BoolFlag")
	assert.False(t, boolFlag.Value, "attach flag should default to false")
	assert.Contains(t, boolFlag.Aliases, "a", "attach flag should expose '-a' shorthand")
	assert.True(t, boolFlag.Local, "attach flag should be local to the run command")
}

func TestRunCmdAttachFlagIsCategorized(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	attachFlag, ok := lookupFlag(runCmd.Flags, "attach").(*cli.BoolFlag)
	assert.True(t, ok, "attach flag should be a *cli.BoolFlag")
	assert.NotEmpty(t, attachFlag.Category, "attach flag should be assigned to a category for help output")
}

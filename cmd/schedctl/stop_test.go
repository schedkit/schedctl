package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cmd "schedctl/cmd/schedctl"
)

func TestNewStopCmd(t *testing.T) {
	stopCmd := cmd.NewStopCmd()

	assert.NotNil(t, stopCmd)
	assert.Equal(t, "stop", stopCmd.Name)
	assert.Equal(t, "stop a scheduler", stopCmd.Usage)
	assert.Equal(t, "ID", stopCmd.ArgsUsage)
}

func TestStopCmdHasNoLocalFlags(t *testing.T) {
	stopCmd := cmd.NewStopCmd()

	assert.Empty(t, stopCmd.Flags, "stop command should rely on the persistent --driver flag, not local ones")
}

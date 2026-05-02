package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cmd "schedctl/cmd/schedctl"
)

func TestNewStatusCmd(t *testing.T) {
	c := cmd.NewStatusCmd()

	assert.NotNil(t, c)
	assert.Equal(t, "status", c.Name)
	assert.NotEmpty(t, c.Usage)
}

func TestStatusCmdHasOutputFlag(t *testing.T) {
	c := cmd.NewStatusCmd()

	outputFlag := lookupFlag(c.Flags, "output")
	assert.NotNil(t, outputFlag, "status command should expose --output")
	assert.Contains(t, outputFlag.Names(), "o", "status --output should expose '-o' shorthand")
}

package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cmd "schedctl/cmd/schedctl"
)

func TestNewPsCmd(t *testing.T) {
	psCmd := cmd.NewPsCmd()

	assert.NotNil(t, psCmd)
	assert.Equal(t, "ps", psCmd.Name)
	assert.Equal(t, "list running schedulers", psCmd.Usage)
}

func TestPsCmdHasNoLocalFlags(t *testing.T) {
	psCmd := cmd.NewPsCmd()

	assert.Empty(t, psCmd.Flags, "ps command should rely on the persistent --driver flag, not local ones")
}

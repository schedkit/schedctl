package cmd_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"schedctl/cmd/schedctl"
)

func TestNewListCmd(t *testing.T) {
	listCmd := cmd.NewListCmd()

	assert.NotNil(t, listCmd)
	assert.Equal(t, "list", listCmd.Use)
	assert.Equal(t, "list available schedulers", listCmd.Short)
}

func TestListCmdExecute(t *testing.T) {
	listCmd := cmd.NewListCmd()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listCmd.Execute()
	w.Close()
	assert.NoError(t, err)

	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)

	assert.NoError(t, err)
	output := buf.String()
	assert.NotEmpty(t, output, "List command should produce output")
}

func TestListCmdHasNoFlags(t *testing.T) {
	listCmd := cmd.NewListCmd()

	assert.False(t, listCmd.HasPersistentFlags(), "List command should not have persistent flags")
	assert.False(t, listCmd.HasLocalFlags(), "List command should not have local flags")
}

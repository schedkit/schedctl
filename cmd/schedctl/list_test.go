package cmd_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	cmd "schedctl/cmd/schedctl"
)

func TestNewListCmd(t *testing.T) {
	listCmd := cmd.NewListCmd()

	assert.NotNil(t, listCmd)
	assert.Equal(t, "list", listCmd.Name)
	assert.Equal(t, "list available schedulers", listCmd.Usage)
}

func TestListCmdExecute(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := rootCmd.Run(context.Background(), []string{"schedctl", "list"})
	w.Close()
	os.Stdout = oldStdout

	assert.NoError(t, err)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	assert.NoError(t, err)
	assert.NotEmpty(t, buf.String(), "List command should produce output")
}

func TestListCmdHasNoLocalFlags(t *testing.T) {
	listCmd := cmd.NewListCmd()

	assert.Empty(t, listCmd.Flags, "List command should declare no flags of its own")
}

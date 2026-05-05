package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cmd "schedctl/cmd/schedctl"
)

func TestNewVersionsCmd(t *testing.T) {
	versionsCmd := cmd.NewVersionsCmd()

	assert.NotNil(t, versionsCmd)
	assert.Equal(t, "versions", versionsCmd.Name)
	assert.Equal(t, "list available versions for a scheduler", versionsCmd.Usage)
	assert.Equal(t, "SCHEDULER", versionsCmd.ArgsUsage)
	assert.NotEmpty(t, versionsCmd.Description)
}

func TestVersionsCmdHasNoLocalFlags(t *testing.T) {
	versionsCmd := cmd.NewVersionsCmd()

	assert.Empty(t, versionsCmd.Flags, "versions command should have no local flags")
}

func TestVersionsCmdRegisteredOnRoot(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	assert.NotNil(t, findSubcommand(rootCmd, "versions"), "versions subcommand must be registered on the root command")
}

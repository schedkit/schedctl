package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"

	cmd "schedctl/cmd/schedctl"
)

func TestNewDoctorCmd(t *testing.T) {
	doctorCmd := cmd.NewDoctorCmd()

	assert.NotNil(t, doctorCmd)
	assert.Equal(t, "doctor", doctorCmd.Name)
	assert.Equal(t, "check host readiness for sched_ext schedulers", doctorCmd.Usage)
	assert.NotEmpty(t, doctorCmd.Description)
}

func TestDoctorCmdHasOutputFlag(t *testing.T) {
	doctorCmd := cmd.NewDoctorCmd()

	outputFlag := lookupFlag(doctorCmd.Flags, "output")
	assert.NotNil(t, outputFlag, "doctor command should have 'output' flag")

	stringFlag, ok := outputFlag.(*cli.StringFlag)
	assert.True(t, ok, "output flag should be a StringFlag")
	assert.Equal(t, "text", stringFlag.Value, "output flag should default to 'text'")
	assert.Contains(t, stringFlag.Aliases, "o", "output flag should expose '-o' shorthand")
	assert.True(t, stringFlag.Local, "output flag should be local to doctor")
	assert.Contains(t, stringFlag.Usage, "json")
}

func TestDoctorCmdRegisteredOnRoot(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	assert.NotNil(t, findSubcommand(rootCmd, "doctor"), "doctor subcommand must be registered on the root command")
}

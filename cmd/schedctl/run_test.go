package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"schedctl/cmd/schedctl"
)

func TestNewRunCmd(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	assert.NotNil(t, runCmd)
	assert.Equal(t, "run SCHEDULER [-- ARGS...]", runCmd.Use)
	assert.Equal(t, "Run a specific scheduler", runCmd.Short)
}

func TestRunCmdHasAttachFlag(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	attachFlag := runCmd.Flags().Lookup("attach")
	assert.NotNil(t, attachFlag, "run command should have 'attach' flag")
	assert.Equal(t, "false", attachFlag.DefValue, "attach flag should default to 'false'")
	assert.Equal(t, "a", attachFlag.Shorthand, "attach flag should have 'a' shorthand")
}

func TestRunCmdHasDriverFlag(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	driverFlag := runCmd.PersistentFlags().Lookup("driver")
	assert.NotNil(t, driverFlag, "run command should have 'driver' flag")
	assert.Equal(t, "podman", driverFlag.DefValue, "driver flag should default to 'podman'")
	assert.Equal(t, "d", driverFlag.Shorthand, "driver flag should have 'd' shorthand")
}

func TestRunCmdDriverFlagUsage(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	driverFlag := runCmd.PersistentFlags().Lookup("driver")
	assert.Contains(t, driverFlag.Usage, "containerd", "driver flag usage should mention 'containerd'")
	assert.Contains(t, driverFlag.Usage, "podman", "driver flag usage should mention 'podman'")
}

func TestRunCmdAttachFlagShorthand(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	err := runCmd.ParseFlags([]string{"-a"})
	assert.NoError(t, err)

	attachFlag := runCmd.Flags().Lookup("attach")
	assert.Equal(t, "true", attachFlag.Value.String())
}

func TestRunCmdAttachFlagLongForm(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	err := runCmd.ParseFlags([]string{"--attach"})
	assert.NoError(t, err)

	attachFlag := runCmd.Flags().Lookup("attach")
	assert.Equal(t, "true", attachFlag.Value.String())
}

func TestRunCmdDriverFlagShorthand(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	err := runCmd.ParseFlags([]string{"-d", "containerd"})
	assert.NoError(t, err)

	driverFlag := runCmd.Flags().Lookup("driver")
	assert.Equal(t, "containerd", driverFlag.Value.String())
}

func TestRunCmdDriverFlagLongForm(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	// Test that the long flag works
	err := runCmd.ParseFlags([]string{"--driver", "containerd"})
	assert.NoError(t, err)

	driverFlag := runCmd.Flags().Lookup("driver")
	assert.Equal(t, "containerd", driverFlag.Value.String())
}

func TestRunCmdDefaultDriverValue(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	driverFlag := runCmd.PersistentFlags().Lookup("driver")
	assert.Equal(t, "podman", driverFlag.Value.String(), "Default driver should be 'podman'")
}

func TestRunCmdDefaultAttachValue(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	attachFlag := runCmd.Flags().Lookup("attach")
	assert.Equal(t, "false", attachFlag.Value.String(), "Default attach should be 'false'")
}

func TestRunCmdAcceptsArgs(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	err := runCmd.ParseFlags([]string{"scheduler-name"})
	assert.NoError(t, err)
}

func TestRunCmdCombinedFlags(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	err := runCmd.ParseFlags([]string{"-a", "-d", "containerd", "scheduler-name"})
	assert.NoError(t, err)

	attachFlag := runCmd.Flags().Lookup("attach")
	driverFlag := runCmd.Flags().Lookup("driver")

	assert.Equal(t, "true", attachFlag.Value.String())
	assert.Equal(t, "containerd", driverFlag.Value.String())
}

func TestRunCmdWithContainerArgs(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	err := runCmd.ParseFlags([]string{"scx_rusty", "--", "--verbose", "--interval=100"})
	assert.NoError(t, err)

	args := runCmd.Flags().Args()
	assert.Equal(t, 3, len(args))
	assert.Equal(t, "scx_rusty", args[0])
	assert.Equal(t, "--verbose", args[1])
	assert.Equal(t, "--interval=100", args[2])
	assert.Equal(t, 1, runCmd.ArgsLenAtDash())
}

func TestRunCmdWithoutContainerArgs(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	err := runCmd.ParseFlags([]string{"scx_rusty"})
	assert.NoError(t, err)

	args := runCmd.Flags().Args()
	assert.Equal(t, 1, len(args))
	assert.Equal(t, "scx_rusty", args[0])
	assert.Equal(t, -1, runCmd.ArgsLenAtDash())
}

func TestRunCmdArgsSeparatorWithFlags(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	err := runCmd.ParseFlags([]string{"-a", "-d", "containerd", "scx_rusty", "--", "--mode=perf"})
	assert.NoError(t, err)

	attachFlag := runCmd.Flags().Lookup("attach")
	driverFlag := runCmd.Flags().Lookup("driver")
	args := runCmd.Flags().Args()

	assert.Equal(t, "true", attachFlag.Value.String())
	assert.Equal(t, "containerd", driverFlag.Value.String())
	assert.Equal(t, 2, len(args))
	assert.Equal(t, 1, runCmd.ArgsLenAtDash())
}

func TestRunCmdValidationNoSchedulerID(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	err := runCmd.ValidateArgs([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one scheduler ID required")
}

func TestRunCmdValidationMultipleSchedulersBeforeDash(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	err := runCmd.ParseFlags([]string{"sched1", "sched2", "--", "--verbose"})
	assert.NoError(t, err)

	err = runCmd.ValidateArgs(runCmd.Flags().Args())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one scheduler ID required before --")
}

func TestRunCmdValidationEmptyContainerArgs(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	err := runCmd.ParseFlags([]string{"scx_rusty", "--"})
	assert.NoError(t, err)
	assert.Equal(t, 1, runCmd.ArgsLenAtDash())

	err = runCmd.ValidateArgs(runCmd.Flags().Args())
	assert.NoError(t, err)
}

func TestRunCmdValidationSingleSchedulerWithArgs(t *testing.T) {
	runCmd := cmd.NewRunCmd()

	err := runCmd.ParseFlags([]string{"scx_rusty", "--", "--verbose", "--interval=100"})
	assert.NoError(t, err)

	err = runCmd.ValidateArgs(runCmd.Flags().Args())
	assert.NoError(t, err)
}

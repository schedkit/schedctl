package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"schedctl/cmd/schedctl"
)

func TestNewStopCmd(t *testing.T) {
	stopCmd := cmd.NewStopCmd()

	assert.NotNil(t, stopCmd)
	assert.Equal(t, "stop", stopCmd.Use)
	assert.Equal(t, "stop a scheduler", stopCmd.Short)
}

func TestStopCmdHasDriverFlag(t *testing.T) {
	stopCmd := cmd.NewStopCmd()

	driverFlag := stopCmd.PersistentFlags().Lookup("driver")
	assert.NotNil(t, driverFlag, "stop command should have 'driver' flag")
	assert.Equal(t, "podman", driverFlag.DefValue, "driver flag should default to 'podman'")
	assert.Equal(t, "d", driverFlag.Shorthand, "driver flag should have 'd' shorthand")
}

func TestStopCmdDriverFlagUsage(t *testing.T) {
	stopCmd := cmd.NewStopCmd()

	driverFlag := stopCmd.PersistentFlags().Lookup("driver")
	assert.Contains(t, driverFlag.Usage, "containerd", "driver flag usage should mention 'containerd'")
	assert.Contains(t, driverFlag.Usage, "podman", "driver flag usage should mention 'podman'")
}

func TestStopCmdFlagShorthand(t *testing.T) {
	stopCmd := cmd.NewStopCmd()

	// Test that the short flag works
	err := stopCmd.ParseFlags([]string{"-d", "containerd"})
	assert.NoError(t, err)

	driverFlag := stopCmd.Flags().Lookup("driver")
	assert.Equal(t, "containerd", driverFlag.Value.String())
}

func TestStopCmdFlagLongForm(t *testing.T) {
	stopCmd := cmd.NewStopCmd()

	// Test that the long flag works
	err := stopCmd.ParseFlags([]string{"--driver", "containerd"})
	assert.NoError(t, err)

	driverFlag := stopCmd.Flags().Lookup("driver")
	assert.Equal(t, "containerd", driverFlag.Value.String())
}

func TestStopCmdDefaultDriverValue(t *testing.T) {
	stopCmd := cmd.NewStopCmd()

	driverFlag := stopCmd.PersistentFlags().Lookup("driver")
	assert.Equal(t, "podman", driverFlag.Value.String(), "Default driver should be 'podman'")
}

func TestStopCmdAcceptsArgs(t *testing.T) {
	stopCmd := cmd.NewStopCmd()

	// The stop command should accept at least one argument (container ID)
	// We're just testing that parsing doesn't fail with args
	err := stopCmd.ParseFlags([]string{"container-id-123"})
	assert.NoError(t, err)
}

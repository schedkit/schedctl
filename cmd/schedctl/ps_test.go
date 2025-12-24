package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"schedctl/cmd/schedctl"
)

func TestNewPsCmd(t *testing.T) {
	psCmd := cmd.NewPsCmd()

	assert.NotNil(t, psCmd)
	assert.Equal(t, "ps", psCmd.Use)
	assert.Equal(t, "list running schedulers", psCmd.Short)
}

func TestPsCmdHasDriverFlag(t *testing.T) {
	psCmd := cmd.NewPsCmd()

	driverFlag := psCmd.PersistentFlags().Lookup("driver")
	assert.NotNil(t, driverFlag, "ps command should have 'driver' flag")
	assert.Equal(t, "podman", driverFlag.DefValue, "driver flag should default to 'podman'")
	assert.Equal(t, "d", driverFlag.Shorthand, "driver flag should have 'd' shorthand")
}

func TestPsCmdDriverFlagUsage(t *testing.T) {
	psCmd := cmd.NewPsCmd()

	driverFlag := psCmd.PersistentFlags().Lookup("driver")
	assert.Contains(t, driverFlag.Usage, "containerd", "driver flag usage should mention 'containerd'")
	assert.Contains(t, driverFlag.Usage, "podman", "driver flag usage should mention 'podman'")
}

func TestPsCmdFlagShorthand(t *testing.T) {
	psCmd := cmd.NewPsCmd()

	// Test that the short flag works
	err := psCmd.ParseFlags([]string{"-d", "containerd"})
	assert.NoError(t, err)

	driverFlag := psCmd.Flags().Lookup("driver")
	assert.Equal(t, "containerd", driverFlag.Value.String())
}

func TestPsCmdFlagLongForm(t *testing.T) {
	psCmd := cmd.NewPsCmd()

	err := psCmd.ParseFlags([]string{"--driver", "containerd"})
	assert.NoError(t, err)

	driverFlag := psCmd.Flags().Lookup("driver")
	assert.Equal(t, "containerd", driverFlag.Value.String())
}

func TestPsCmdDefaultDriverValue(t *testing.T) {
	psCmd := cmd.NewPsCmd()

	driverFlag := psCmd.PersistentFlags().Lookup("driver")
	assert.Equal(t, "podman", driverFlag.Value.String(), "Default driver should be 'podman'")
}

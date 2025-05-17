package podman_test

import (
	"os/user"
	"testing"

	"github.com/stretchr/testify/assert"

	"schedctl/internal/podman"
)

func TestPodmanSpawnStopProcesses(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}

	if u.Username != "root" {
		t.SkipNow()
	}

	containers, err := podman.List()
	assert.Nil(t, err)

	assert.Equal(t, 0, len(containers))

	err = podman.Run("ghcr.io/schedkit/scx_rusty:latest", "test-scheduler")
	assert.Nil(t, err)

	containers, err = podman.List()
	assert.Nil(t, err)

	assert.Equal(t, 1, len(containers))

	err = podman.Stop("test-scheduler")
	assert.Nil(t, err)
}

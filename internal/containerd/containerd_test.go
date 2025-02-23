package containerd_test

import (
	"schedctl/internal/containerd"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerdSpawnStopProcess(t *testing.T) {
	client, err := containerd.NewClient()
	assert.Nil(t, err)

	containers, err := containerd.List(client)
	assert.Nil(t, err)

	assert.Equal(t, 0, len(containers))

	err = containerd.Run(client, "ghcr.io/schedkit/scx_rusty:latest", "test-scheduler", false)
	assert.Nil(t, err)

	containers, err = containerd.List(client)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(containers))

	err = containerd.Stop(client, "test-scheduler")
	assert.Nil(t, err)
}

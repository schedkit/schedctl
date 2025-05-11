package containerd_test

import (
	"os/user"
	"testing"

	"github.com/stretchr/testify/assert"

	"schedctl/internal/containerd"
)

func TestContainerdSpawnStopProcess(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}

	if u.Username != "root" {
		t.SkipNow()
	}

	client, err := containerd.NewClient()
	assert.Nil(t, err)

	containers, err := containerd.List(client)
	assert.Nil(t, err)

	assert.Equal(t, 0, len(containers))

	err = containerd.Run(client, "ghcr.io/schedkit/scx_rusty:latest", "test-scheduler", false, true)
	assert.Nil(t, err)

	containers, err = containerd.List(client)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(containers))

	err = containerd.Stop(client, "test-scheduler")
	assert.Nil(t, err)
}

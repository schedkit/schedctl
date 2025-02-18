package schedulers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"schedctl/internal/schedulers"
)

func TestSchedulerFound(t *testing.T) {
	schedulerImage, err := schedulers.GetScheduler("scx_rusty")
	assert.Nil(t, err)
	assert.Equal(t, "ghcr.io/schedkit/scx_rusty:latest", schedulerImage)
}

func TestSchedulerNotFound(t *testing.T) {
	_, err := schedulers.GetScheduler("unknown")
	assert.NotNil(t, err)
}

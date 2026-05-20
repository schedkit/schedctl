package schedulers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"schedctl/internal/schedulers"
)

const rustyRepo = "ghcr.io/schedkit/scx_rusty"

func TestSchedulerFound(t *testing.T) {
	result, err := schedulers.GetScheduler("scx_rusty", "")
	assert.Nil(t, err)
	assert.Equal(t, rustyRepo+":latest", result.ImageURI)
	assert.Equal(t, schedulers.SourceManifest, result.Source)
}

func TestSchedulerNotFound(t *testing.T) {
	_, err := schedulers.GetScheduler("unknown", "")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "scheduler not found")
}

func TestDirectImageWithRegistry(t *testing.T) {
	result, err := schedulers.GetScheduler("ghcr.io/myorg/custom-scheduler", "")
	assert.Nil(t, err)
	assert.Equal(t, "ghcr.io/myorg/custom-scheduler:latest", result.ImageURI)
	assert.Equal(t, schedulers.SourceDirect, result.Source)
}

func TestDirectImageWithTag(t *testing.T) {
	result, err := schedulers.GetScheduler("docker.io/nginx:alpine", "")
	assert.Nil(t, err)
	assert.Equal(t, "docker.io/nginx:alpine", result.ImageURI)
	assert.Equal(t, schedulers.SourceDirect, result.Source)
}

func TestDirectImageWithDigest(t *testing.T) {
	result, err := schedulers.GetScheduler("quay.io/org/image@sha256:abc123", "")
	assert.Nil(t, err)
	assert.Equal(t, "quay.io/org/image@sha256:abc123", result.ImageURI)
	assert.Equal(t, schedulers.SourceDirect, result.Source)
}

func TestDirectImageLocalhost(t *testing.T) {
	result, err := schedulers.GetScheduler("localhost:5000/myimage", "")
	assert.Nil(t, err)
	assert.Equal(t, "localhost:5000/myimage:latest", result.ImageURI)
	assert.Equal(t, schedulers.SourceDirect, result.Source)
}

func TestDirectImageDockerHub(t *testing.T) {
	result, err := schedulers.GetScheduler("nginx/alpine", "")
	assert.Nil(t, err)
	assert.Equal(t, "nginx/alpine:latest", result.ImageURI)
	assert.Equal(t, schedulers.SourceDirect, result.Source)
}

func TestInvalidInput(t *testing.T) {
	_, err := schedulers.GetScheduler("just-a-name", "")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "scheduler not found")
}

func TestGetSchedulerWithVersion(t *testing.T) {
	result, err := schedulers.GetScheduler("scx_rusty", "v1.0.0")
	assert.Nil(t, err)
	assert.Equal(t, rustyRepo+":v1.0.0", result.ImageURI)
	assert.Equal(t, schedulers.SourceManifest, result.Source)
}

func TestGetSchedulerWithVersionOverridesExistingTag(t *testing.T) {
	result, err := schedulers.GetScheduler("docker.io/nginx:alpine", "v2.0")
	assert.Nil(t, err)
	assert.Equal(t, "docker.io/nginx:v2.0", result.ImageURI)
	assert.Equal(t, schedulers.SourceDirect, result.Source)
}

func TestImageRepo(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{rustyRepo + ":latest", rustyRepo},
		{rustyRepo, rustyRepo},
		{rustyRepo + "@sha256:abc123", rustyRepo},
		{"localhost:5000/myimage:v1", "localhost:5000/myimage"},
		{"docker.io/nginx:alpine", "docker.io/nginx"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, schedulers.ImageRepo(tt.input))
		})
	}
}

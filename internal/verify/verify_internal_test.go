package verify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepoFromRefStripsTagAndDigest(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"ghcr.io/schedkit/scx_rusty:latest", "ghcr.io/schedkit/scx_rusty"},
		{"ghcr.io/schedkit/scx_rusty", "ghcr.io/schedkit/scx_rusty"},
		{"ghcr.io/schedkit/scx_rusty@sha256:abc123", "ghcr.io/schedkit/scx_rusty"},
		{"docker.io/library/nginx:1.25", "index.docker.io/library/nginx"},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := repoFromRef(c.in)
			assert.NoError(t, err)
			assert.Equal(t, c.want, got)
		})
	}
}

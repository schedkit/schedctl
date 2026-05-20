package verify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const rustyRepo = "ghcr.io/schedkit/scx_rusty"

func TestRepoFromRefStripsTagAndDigest(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{rustyRepo + ":latest", rustyRepo},
		{rustyRepo, rustyRepo},
		{rustyRepo + "@sha256:abc123", rustyRepo},
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

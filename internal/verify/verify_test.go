package verify_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"schedctl/internal/verify"
)

func TestSignerDescribeKeyless(t *testing.T) {
	s := verify.Signer{
		Issuer:  "https://token.actions.githubusercontent.com",
		Subject: "https://github.com/schedkit/plumbing/.github/workflows/release.yaml@refs/heads/main",
	}
	got := s.Describe()
	assert.Contains(t, got, "keyless")
	assert.Contains(t, got, "github.com/schedkit/plumbing")
	assert.Contains(t, got, "githubusercontent.com")
}

func TestSignerDescribeKey(t *testing.T) {
	s := verify.Signer{
		KeyPath:        "/etc/schedctl/keys/release.pub",
		KeyFingerprint: "abcd1234",
	}
	got := s.Describe()
	assert.Contains(t, got, "key /etc/schedctl/keys/release.pub")
	assert.Contains(t, got, "sha256:abcd1234")
}

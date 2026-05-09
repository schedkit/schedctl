//go:build e2e

package verify_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"schedctl/internal/verify"
)

const (
	signedSchedkitImage = "ghcr.io/schedkit/scx_rusty:latest"
	unsignedImage       = "docker.io/library/alpine:3.19"
)

// TestE2EVerifySchedkitImage exercises the full cosign verification path
// (registry + Fulcio roots + Rekor + CTLog pubs) against a real signed
// schedkit image using the built-in default policy. Requires network.
func TestE2EVerifySchedkitImage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	policy := verify.DefaultPolicy()
	result, err := verify.Image(ctx, signedSchedkitImage, policy)
	require.NoError(t, err, "default policy should accept ghcr.io/schedkit/scx_rusty")

	assert.True(t, strings.HasPrefix(result.Digest, "sha256:"),
		"expected sha256 digest, got %q", result.Digest)
	assert.Contains(t, result.ImageRef, "@"+result.Digest,
		"verified image ref should be digest-pinned")

	require.Len(t, result.Signers, 1)
	signer := result.Signers[0]
	assert.Equal(t, "https://token.actions.githubusercontent.com", signer.Issuer)
	assert.Contains(t, signer.Subject, "github.com/schedkit",
		"signer subject should be a schedkit org workflow, got %q", signer.Subject)

	t.Logf("verified %s signed by %s", result.ImageRef, signer.Describe())
}

// TestE2EVerifyUnsignedImageRejected confirms that an unsigned image is a
// hard failure under the default policy.
func TestE2EVerifyUnsignedImageRejected(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	policy := verify.DefaultPolicy()
	_, err := verify.Image(ctx, unsignedImage, policy)
	require.Error(t, err, "unsigned %q must not pass verification", unsignedImage)
}

// TestE2EVerifyWrongIdentityRejected confirms a signed image fails when the
// policy demands a different OIDC identity.
func TestE2EVerifyWrongIdentityRejected(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	policy := verify.Policy{
		Identities: []verify.Identity{{
			Issuer:        "https://token.actions.githubusercontent.com",
			SubjectRegExp: `^https://github\.com/some-other-org/.*`,
		}},
		RekorURL: "https://rekor.sigstore.dev",
	}
	_, err := verify.Image(ctx, signedSchedkitImage, policy)
	require.Error(t, err, "policy targeting a different org must reject schedkit signature")
}

package verify_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"schedctl/internal/verify"
)

func TestLoadPolicyDefault(t *testing.T) {
	t.Setenv("SCHEDCTL_TRUST_POLICY", "")

	p, err := verify.LoadPolicy("")
	require.NoError(t, err)

	assert.Empty(t, p.Keys)
	assert.NotEmpty(t, p.Identities, "default policy should ship at least one identity")
	assert.Equal(t, "https://token.actions.githubusercontent.com", p.Identities[0].Issuer)
	assert.Contains(t, p.Identities[0].SubjectRegExp, "schedkit")
	assert.Equal(t, "https://rekor.sigstore.dev", p.RekorURL)
}

func TestLoadPolicyFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`
identities:
  - issuer: https://accounts.google.com
    subject: releases@example.com
rekorURL: https://rekor.example.com
`), 0o600))

	p, err := verify.LoadPolicy(path)
	require.NoError(t, err)

	require.Len(t, p.Identities, 1)
	assert.Equal(t, "https://accounts.google.com", p.Identities[0].Issuer)
	assert.Equal(t, "releases@example.com", p.Identities[0].Subject)
	assert.Equal(t, "https://rekor.example.com", p.RekorURL)
}

func TestLoadPolicyFromEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`
keys:
  - path: /etc/keys/release.pub
`), 0o600))

	t.Setenv("SCHEDCTL_TRUST_POLICY", path)

	p, err := verify.LoadPolicy("")
	require.NoError(t, err)

	require.Len(t, p.Keys, 1)
	assert.Equal(t, "/etc/keys/release.pub", p.Keys[0].Path)
}

func TestLoadPolicyExplicitBeatsEnv(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, "env.yaml")
	require.NoError(t, os.WriteFile(envPath, []byte("identities:\n  - issuer: env\n"), 0o600))
	flagPath := filepath.Join(dir, "flag.yaml")
	require.NoError(t, os.WriteFile(flagPath, []byte("identities:\n  - issuer: flag\n"), 0o600))

	t.Setenv("SCHEDCTL_TRUST_POLICY", envPath)

	p, err := verify.LoadPolicy(flagPath)
	require.NoError(t, err)
	require.Len(t, p.Identities, 1)
	assert.Equal(t, "flag", p.Identities[0].Issuer)
}

func TestLoadPolicyMissingFile(t *testing.T) {
	_, err := verify.LoadPolicy("/nonexistent/policy.yaml")
	assert.Error(t, err)
}

func TestLoadPolicyEmptyDocument(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	require.NoError(t, os.WriteFile(path, []byte("rekorURL: https://rekor.example.com\n"), 0o600))

	_, err := verify.LoadPolicy(path)
	assert.Error(t, err, "policy with no keys or identities should be rejected")
}

func TestLoadPolicyInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.yaml")
	require.NoError(t, os.WriteFile(path, []byte("not: [valid"), 0o600))

	_, err := verify.LoadPolicy(path)
	assert.Error(t, err)
}

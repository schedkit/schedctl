package verify

import (
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

const trustPolicyEnv = "SCHEDCTL_TRUST_POLICY"

type Key struct {
	Path string `json:"path"`
}

type Identity struct {
	Issuer        string `json:"issuer"`
	Subject       string `json:"subject"`
	IssuerRegExp  string `json:"issuerRegExp"`
	SubjectRegExp string `json:"subjectRegExp"`
}

type Policy struct {
	Keys       []Key      `json:"keys"`
	Identities []Identity `json:"identities"`
	RekorURL   string     `json:"rekorURL"`
}

// LoadPolicy resolves a trust policy from (in order): the explicit path, the
// SCHEDCTL_TRUST_POLICY env var, or the built-in default. An empty path with no
// env var falls through to DefaultPolicy.
func LoadPolicy(path string) (Policy, error) {
	if path == "" {
		path = os.Getenv(trustPolicyEnv)
	}
	if path == "" {
		return DefaultPolicy(), nil
	}

	raw, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return Policy{}, fmt.Errorf("read trust policy %q: %w", path, err)
	}

	var p Policy
	if err := yaml.Unmarshal(raw, &p); err != nil {
		return Policy{}, fmt.Errorf("parse trust policy %q: %w", path, err)
	}
	if len(p.Keys) == 0 && len(p.Identities) == 0 {
		return Policy{}, fmt.Errorf("trust policy %q has no keys or identities", path)
	}
	return p, nil
}

// DefaultPolicy returns the built-in policy that trusts keyless cosign
// signatures from the schedkit GitHub org's GitHub Actions workflows.
func DefaultPolicy() Policy {
	return Policy{
		Identities: []Identity{{
			Issuer:        "https://token.actions.githubusercontent.com",
			SubjectRegExp: `^https://github\.com/schedkit/.*`,
		}},
		RekorURL: "https://rekor.sigstore.dev",
	}
}

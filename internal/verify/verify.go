package verify

import (
	"context"
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sigstore/cosign/v2/pkg/oci"
	sgverify "github.com/sigstore/sigstore-go/pkg/verify"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/sigstore/sigstore/pkg/fulcioroots"
	"github.com/sigstore/sigstore/pkg/signature"
)

type Signer struct {
	Issuer         string
	Subject        string
	KeyPath        string
	KeyFingerprint string
}

func (s Signer) Describe() string {
	if s.KeyPath != "" {
		return fmt.Sprintf("key %s (sha256:%s)", s.KeyPath, s.KeyFingerprint)
	}
	return fmt.Sprintf("keyless %s from %s", s.Subject, s.Issuer)
}

type Result struct {
	ImageRef string
	Digest   string
	Signers  []Signer
	Skipped  bool
}

func Image(ctx context.Context, imageRef string, policy Policy) (Result, error) {
	digest, err := crane.Digest(imageRef)
	if err != nil {
		return Result{}, fmt.Errorf("resolve digest for %q: %w", imageRef, err)
	}
	repo, err := repoFromRef(imageRef)
	if err != nil {
		return Result{}, err
	}
	pinned := repo + "@" + digest
	digestRef, err := name.NewDigest(pinned)
	if err != nil {
		return Result{}, fmt.Errorf("build digest reference: %w", err)
	}

	if len(policy.Keys) > 0 {
		signer, err := verifyByKeys(ctx, digestRef, policy)
		if err == nil {
			return Result{ImageRef: pinned, Digest: digest, Signers: []Signer{signer}}, nil
		}
		if len(policy.Identities) == 0 {
			return Result{}, fmt.Errorf("verify against trusted keys: %w", err)
		}
	}

	if len(policy.Identities) > 0 {
		signer, err := verifyByIdentities(ctx, digestRef, policy)
		if err != nil {
			return Result{}, fmt.Errorf("verify against trusted identities: %w", err)
		}
		return Result{ImageRef: pinned, Digest: digest, Signers: []Signer{signer}}, nil
	}

	return Result{}, errors.New("trust policy has no keys or identities")
}

func verifyByKeys(ctx context.Context, ref name.Digest, policy Policy) (Signer, error) {
	var lastErr error
	for _, k := range policy.Keys {
		verifier, err := signature.LoadVerifierFromPEMFile(k.Path, crypto.SHA256)
		if err != nil {
			lastErr = fmt.Errorf("load key %q: %w", k.Path, err)
			continue
		}
		opts := &cosign.CheckOpts{
			SigVerifier: verifier,
			IgnoreTlog:  true, // key-based signatures don't require Rekor
		}
		if _, _, err := cosign.VerifyImageSignatures(ctx, ref, opts); err != nil {
			lastErr = err
			continue
		}
		fp, err := keyFingerprint(k.Path)
		if err != nil {
			fp = "unknown"
		}
		return Signer{KeyPath: k.Path, KeyFingerprint: fp}, nil
	}
	if lastErr == nil {
		lastErr = errors.New("no trusted keys matched")
	}
	return Signer{}, lastErr
}

func verifyByIdentities(ctx context.Context, ref name.Digest, policy Policy) (Signer, error) {
	identities := make([]cosign.Identity, 0, len(policy.Identities))
	for _, id := range policy.Identities {
		identities = append(identities, cosign.Identity{
			Issuer:        id.Issuer,
			Subject:       id.Subject,
			IssuerRegExp:  id.IssuerRegExp,
			SubjectRegExp: id.SubjectRegExp,
		})
	}

	co := &cosign.CheckOpts{Identities: identities}
	if trusted, err := cosign.TrustedRoot(); err == nil {
		co.TrustedMaterial = trusted
	} else {
		// Fall back to discrete trust material when TUF is unavailable. The
		// new-bundle format requires TrustedMaterial and will fail below
		// with a clear message.
		roots, rerr := fulcioroots.Get()
		if rerr != nil {
			return Signer{}, fmt.Errorf("trusted root unavailable (fulcio fallback also failed): %w", errors.Join(err, rerr))
		}
		co.RootCerts = roots
		if intermediates, ierr := fulcioroots.GetIntermediates(); ierr == nil {
			co.IntermediateCerts = intermediates
		}
		rekorPubs, perr := cosign.GetRekorPubs(ctx)
		if perr != nil {
			return Signer{}, fmt.Errorf("rekor public keys: %w", perr)
		}
		co.RekorPubKeys = rekorPubs
		ctPubs, cerr := cosign.GetCTLogPubs(ctx)
		if cerr != nil {
			return Signer{}, fmt.Errorf("ctlog public keys: %w", cerr)
		}
		co.CTLogPubKeys = ctPubs
	}

	// Prefer the new Sigstore bundle format when present (OCI 1.1 referrer
	// of mediaType vnd.dev.sigstore.bundle.v0.3+json), fall back to the
	// legacy `.sig` tag layout otherwise.
	if bundles, hash, berr := cosign.GetBundles(ctx, ref, co); berr == nil && len(bundles) > 0 {
		if co.TrustedMaterial == nil {
			return Signer{}, errors.New("new bundle format requires sigstore TrustedRoot")
		}
		digestBytes, derr := hex.DecodeString(hash.Hex)
		if derr != nil {
			return Signer{}, fmt.Errorf("decode digest: %w", derr)
		}
		policyOpt := sgverify.WithArtifactDigest(hash.Algorithm, digestBytes)
		var lastErr error
		for _, b := range bundles {
			res, verr := cosign.VerifyNewBundle(ctx, co, policyOpt, b)
			if verr != nil {
				lastErr = verr
				continue
			}
			return signerFromVerificationResult(res), nil
		}
		if lastErr == nil {
			lastErr = errors.New("no bundle verified")
		}
		return Signer{}, lastErr
	}

	co.NewBundleFormat = false
	sigs, _, verr := cosign.VerifyImageSignatures(ctx, ref, co)
	if verr != nil {
		return Signer{}, verr
	}
	if len(sigs) == 0 {
		return Signer{}, errors.New("no valid signatures")
	}
	return signerFromLegacySignature(sigs[0])
}

func signerFromVerificationResult(res *sgverify.VerificationResult) Signer {
	if res == nil {
		return Signer{}
	}
	// Prefer the actual cert's data so we report what was IN the bundle, not
	// what the policy matcher requested (which may be a regex pattern).
	if res.Signature != nil && res.Signature.Certificate != nil {
		return Signer{
			Issuer:  res.Signature.Certificate.Issuer,
			Subject: res.Signature.Certificate.SubjectAlternativeName,
		}
	}
	if res.VerifiedIdentity != nil {
		return Signer{
			Issuer:  res.VerifiedIdentity.Issuer.Issuer,
			Subject: res.VerifiedIdentity.SubjectAlternativeName.SubjectAlternativeName,
		}
	}
	return Signer{}
}

func signerFromLegacySignature(sig oci.Signature) (Signer, error) {
	cert, err := sig.Cert()
	if err != nil || cert == nil {
		return Signer{}, fmt.Errorf("read signing cert: %w", err)
	}
	subjects := cryptoutils.GetSubjectAlternateNames(cert)
	subject := ""
	if len(subjects) > 0 {
		subject = subjects[0]
	}
	return Signer{Issuer: oidcIssuerFromCert(cert), Subject: subject}, nil
}

func oidcIssuerFromCert(cert *x509.Certificate) string {
	oidIssuer := []int{1, 3, 6, 1, 4, 1, 57264, 1, 1}
	oidIssuerV2DER := []int{1, 3, 6, 1, 4, 1, 57264, 1, 8}
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(oidIssuerV2DER) && len(ext.Value) > 2 {
			return string(ext.Value[2:])
		}
	}
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(oidIssuer) {
			return string(ext.Value)
		}
	}
	return ""
}

func repoFromRef(ref string) (string, error) {
	if i := strings.Index(ref, "@"); i != -1 {
		return ref[:i], nil
	}
	parsed, err := name.ParseReference(ref)
	if err != nil {
		return "", fmt.Errorf("parse image reference %q: %w", ref, err)
	}
	return parsed.Context().Name(), nil
}

func keyFingerprint(path string) (string, error) {
	verifier, err := signature.LoadVerifierFromPEMFile(path, crypto.SHA256)
	if err != nil {
		return "", err
	}
	pub, err := verifier.PublicKey()
	if err != nil {
		return "", err
	}
	der, err := cryptoutils.MarshalPublicKeyToDER(pub)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(der)
	return hex.EncodeToString(sum[:8]), nil
}

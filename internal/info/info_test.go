package info_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"schedctl/internal/info"
	"schedctl/internal/schedulers"
)

const (
	rustyImage       = "ghcr.io/schedkit/scx_rusty:latest"
	rustyDigest      = "sha256:abc"
	rustyName        = "scx_rusty"
	rustyDescription = "Rust user-space scheduler"
	rustyDocs        = "https://example.com/docs"
	rustySource      = "https://github.com/sched-ext/scx"
	rustyVersion     = "1.0.0"
	rustyKernelMin   = "6.12"
	sparseImage      = "ghcr.io/example/sparse:latest"
)

func TestBuildPopulatesAllCuratedFields(t *testing.T) {
	created := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	img := schedulers.SchedulerImage{
		ImageURI: rustyImage,
		Source:   schedulers.SourceManifest,
	}
	raw := schedulers.ImageInfo{
		ImageRef:     rustyImage,
		Digest:       rustyDigest,
		Created:      &created,
		Architecture: "amd64",
		OS:           "linux",
		Labels: map[string]string{
			info.LabelTitle:         rustyName,
			info.LabelDescription:   rustyDescription,
			info.LabelDocumentation: rustyDocs,
			info.LabelSource:        rustySource,
			info.LabelVersion:       rustyVersion,
			info.LabelRevision:      "deadbee",
			info.LabelAuthors:       "scx team",
			info.LabelLicenses:      "GPL-2.0",
			info.LabelVendor:        "schedkit",
			info.LabelKernelMin:     rustyKernelMin,
		},
	}

	r := info.Build(rustyName, img, raw)

	assert.Equal(t, info.SchemaVersion, r.SchemaVersion)
	assert.Equal(t, rustyName, r.Scheduler)
	assert.Equal(t, "manifest", r.Source)
	assert.Equal(t, rustyImage, r.ImageRef)
	assert.Equal(t, rustyDigest, r.Digest)
	assert.Equal(t, &created, r.Created)
	assert.Equal(t, "amd64", r.Architecture)
	assert.Equal(t, "linux", r.OS)
	assert.Equal(t, rustyName, r.Title)
	assert.Equal(t, rustyDescription, r.Description)
	assert.Equal(t, rustyDocs, r.Documentation)
	assert.Equal(t, rustySource, r.SourceURL)
	assert.Equal(t, rustyVersion, r.Version)
	assert.Equal(t, "deadbee", r.Revision)
	assert.Equal(t, "scx team", r.Authors)
	assert.Equal(t, "GPL-2.0", r.Licenses)
	assert.Equal(t, "schedkit", r.Vendor)
	assert.Equal(t, rustyKernelMin, r.KernelMin)
}

func TestBuildOmitsMissingLabels(t *testing.T) {
	img := schedulers.SchedulerImage{
		ImageURI: sparseImage,
		Source:   schedulers.SourceDirect,
	}
	raw := schedulers.ImageInfo{
		ImageRef: sparseImage,
		Labels: map[string]string{
			info.LabelTitle: "sparse",
		},
	}

	r := info.Build("ghcr.io/example/sparse", img, raw)

	assert.Equal(t, "direct", r.Source)
	assert.Equal(t, "sparse", r.Title)
	assert.Empty(t, r.Description)
	assert.Empty(t, r.Documentation)
	assert.Empty(t, r.KernelMin)
	assert.Empty(t, r.Authors)
}

func TestBuildWithEmptyLabelMap(t *testing.T) {
	img := schedulers.SchedulerImage{
		ImageURI: "ghcr.io/example/bare:latest",
		Source:   schedulers.SourceDirect,
	}
	raw := schedulers.ImageInfo{
		ImageRef: "ghcr.io/example/bare:latest",
		Labels:   map[string]string{},
	}

	r := info.Build("ghcr.io/example/bare", img, raw)

	assert.Equal(t, info.SchemaVersion, r.SchemaVersion)
	assert.Equal(t, "ghcr.io/example/bare:latest", r.ImageRef)
	assert.Empty(t, r.Title)
}

func TestWriteJSONOmitsEmptyFields(t *testing.T) {
	r := info.Report{
		SchemaVersion: info.SchemaVersion,
		Scheduler:     rustyName,
		Source:        "manifest",
		ImageRef:      rustyImage,
		Digest:        rustyDigest,
		Title:         rustyName,
	}
	var buf bytes.Buffer
	err := info.WriteJSON(&buf, r)
	assert.NoError(t, err)

	var decoded map[string]any
	assert.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))
	assert.Equal(t, "1", decoded["schema_version"])
	assert.Equal(t, rustyName, decoded["scheduler"])
	assert.Equal(t, rustyDigest, decoded["digest"])
	_, hasEmpty := decoded["description"]
	assert.False(t, hasEmpty, "empty fields should be omitted from JSON")
	_, hasKernel := decoded["kernel_min_version"]
	assert.False(t, hasKernel, "empty kernel_min_version should be omitted")
}

func TestWriteTextIncludesKeyFields(t *testing.T) {
	r := info.Report{
		Scheduler:     rustyName,
		Title:         rustyName,
		Description:   rustyDescription,
		ImageRef:      rustyImage,
		Digest:        rustyDigest,
		KernelMin:     rustyKernelMin,
		Documentation: rustyDocs,
		SourceURL:     rustySource,
		Version:       rustyVersion,
	}
	var buf bytes.Buffer
	info.WriteText(&buf, r)
	out := buf.String()

	for _, want := range []string{
		rustyName,
		rustyDescription,
		rustyImage,
		rustyDigest,
		rustyKernelMin,
		rustyDocs,
		rustySource,
		rustyVersion,
	} {
		assert.True(t, strings.Contains(out, want), "expected text output to contain %q", want)
	}
}

func TestWriteTextFallsBackToSchedulerArgWhenTitleMissing(t *testing.T) {
	r := info.Report{
		Scheduler: "ghcr.io/example/sparse",
		ImageRef:  "ghcr.io/example/sparse:latest",
	}
	var buf bytes.Buffer
	info.WriteText(&buf, r)
	assert.True(t, strings.Contains(buf.String(), "ghcr.io/example/sparse"))
}

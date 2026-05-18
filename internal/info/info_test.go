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

func TestBuildPopulatesAllCuratedFields(t *testing.T) {
	created := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	img := schedulers.SchedulerImage{
		ImageURI: "ghcr.io/schedkit/scx_rusty:latest",
		Source:   schedulers.SourceManifest,
	}
	raw := schedulers.ImageInfo{
		ImageRef:     "ghcr.io/schedkit/scx_rusty:latest",
		Digest:       "sha256:abc",
		Created:      &created,
		Architecture: "amd64",
		OS:           "linux",
		Labels: map[string]string{
			info.LabelTitle:         "scx_rusty",
			info.LabelDescription:   "Rust user-space scheduler",
			info.LabelDocumentation: "https://example.com/docs",
			info.LabelSource:        "https://github.com/sched-ext/scx",
			info.LabelVersion:       "1.0.0",
			info.LabelRevision:      "deadbee",
			info.LabelAuthors:       "scx team",
			info.LabelLicenses:      "GPL-2.0",
			info.LabelVendor:        "schedkit",
			info.LabelKernelMin:     "6.12",
		},
	}

	r := info.Build("scx_rusty", img, raw)

	assert.Equal(t, info.SchemaVersion, r.SchemaVersion)
	assert.Equal(t, "scx_rusty", r.Scheduler)
	assert.Equal(t, "manifest", r.Source)
	assert.Equal(t, "ghcr.io/schedkit/scx_rusty:latest", r.ImageRef)
	assert.Equal(t, "sha256:abc", r.Digest)
	assert.Equal(t, &created, r.Created)
	assert.Equal(t, "amd64", r.Architecture)
	assert.Equal(t, "linux", r.OS)
	assert.Equal(t, "scx_rusty", r.Title)
	assert.Equal(t, "Rust user-space scheduler", r.Description)
	assert.Equal(t, "https://example.com/docs", r.Documentation)
	assert.Equal(t, "https://github.com/sched-ext/scx", r.SourceURL)
	assert.Equal(t, "1.0.0", r.Version)
	assert.Equal(t, "deadbee", r.Revision)
	assert.Equal(t, "scx team", r.Authors)
	assert.Equal(t, "GPL-2.0", r.Licenses)
	assert.Equal(t, "schedkit", r.Vendor)
	assert.Equal(t, "6.12", r.KernelMin)
}

func TestBuildOmitsMissingLabels(t *testing.T) {
	img := schedulers.SchedulerImage{
		ImageURI: "ghcr.io/example/sparse:latest",
		Source:   schedulers.SourceDirect,
	}
	raw := schedulers.ImageInfo{
		ImageRef: "ghcr.io/example/sparse:latest",
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
		Scheduler:     "scx_rusty",
		Source:        "manifest",
		ImageRef:      "ghcr.io/schedkit/scx_rusty:latest",
		Digest:        "sha256:abc",
		Title:         "scx_rusty",
	}
	var buf bytes.Buffer
	err := info.WriteJSON(&buf, r)
	assert.NoError(t, err)

	var decoded map[string]any
	assert.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))
	assert.Equal(t, "1", decoded["schema_version"])
	assert.Equal(t, "scx_rusty", decoded["scheduler"])
	assert.Equal(t, "sha256:abc", decoded["digest"])
	_, hasEmpty := decoded["description"]
	assert.False(t, hasEmpty, "empty fields should be omitted from JSON")
	_, hasKernel := decoded["kernel_min_version"]
	assert.False(t, hasKernel, "empty kernel_min_version should be omitted")
}

func TestWriteTextIncludesKeyFields(t *testing.T) {
	r := info.Report{
		Scheduler:     "scx_rusty",
		Title:         "scx_rusty",
		Description:   "Rust user-space scheduler",
		ImageRef:      "ghcr.io/schedkit/scx_rusty:latest",
		Digest:        "sha256:abc",
		KernelMin:     "6.12",
		Documentation: "https://example.com/docs",
		SourceURL:     "https://github.com/sched-ext/scx",
		Version:       "1.0.0",
	}
	var buf bytes.Buffer
	info.WriteText(&buf, r)
	out := buf.String()

	for _, want := range []string{
		"scx_rusty",
		"Rust user-space scheduler",
		"ghcr.io/schedkit/scx_rusty:latest",
		"sha256:abc",
		"6.12",
		"https://example.com/docs",
		"https://github.com/sched-ext/scx",
		"1.0.0",
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

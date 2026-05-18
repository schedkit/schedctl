package info

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"schedctl/internal/schedulers"
)

const SchemaVersion = "1"

const (
	LabelTitle         = "org.opencontainers.image.title"
	LabelDescription   = "org.opencontainers.image.description"
	LabelDocumentation = "org.opencontainers.image.documentation"
	LabelSource        = "org.opencontainers.image.source"
	LabelVersion       = "org.opencontainers.image.version"
	LabelRevision      = "org.opencontainers.image.revision"
	LabelAuthors       = "org.opencontainers.image.authors"
	LabelLicenses      = "org.opencontainers.image.licenses"
	LabelVendor        = "org.opencontainers.image.vendor"
	LabelKernelMin     = "org.schedkit.scheduler.kernel.min-version"
)

type Report struct {
	SchemaVersion string     `json:"schema_version"`
	Scheduler     string     `json:"scheduler"`
	Source        string     `json:"source"`
	ImageRef      string     `json:"image_ref"`
	Digest        string     `json:"digest,omitempty"`
	Created       *time.Time `json:"created,omitempty"`
	Architecture  string     `json:"architecture,omitempty"`
	OS            string     `json:"os,omitempty"`
	Title         string     `json:"title,omitempty"`
	Description   string     `json:"description,omitempty"`
	Documentation string     `json:"documentation,omitempty"`
	SourceURL     string     `json:"source_url,omitempty"`
	Version       string     `json:"version,omitempty"`
	Revision      string     `json:"revision,omitempty"`
	Authors       string     `json:"authors,omitempty"`
	Licenses      string     `json:"licenses,omitempty"`
	Vendor        string     `json:"vendor,omitempty"`
	KernelMin     string     `json:"kernel_min_version,omitempty"`
}

func Build(arg string, img schedulers.SchedulerImage, raw schedulers.ImageInfo) Report {
	return Report{
		SchemaVersion: SchemaVersion,
		Scheduler:     arg,
		Source:        string(img.Source),
		ImageRef:      raw.ImageRef,
		Digest:        raw.Digest,
		Created:       raw.Created,
		Architecture:  raw.Architecture,
		OS:            raw.OS,
		Title:         raw.Labels[LabelTitle],
		Description:   raw.Labels[LabelDescription],
		Documentation: raw.Labels[LabelDocumentation],
		SourceURL:     raw.Labels[LabelSource],
		Version:       raw.Labels[LabelVersion],
		Revision:      raw.Labels[LabelRevision],
		Authors:       raw.Labels[LabelAuthors],
		Licenses:      raw.Labels[LabelLicenses],
		Vendor:        raw.Labels[LabelVendor],
		KernelMin:     raw.Labels[LabelKernelMin],
	}
}

func WriteJSON(w io.Writer, r Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

func WriteText(w io.Writer, r Report) {
	heading := r.Title
	if heading == "" {
		heading = r.Scheduler
	}
	fmt.Fprintln(w, heading)
	if r.Description != "" {
		fmt.Fprintf(w, "  %s\n", r.Description)
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "  image:        %s\n", r.ImageRef)
	if r.Digest != "" {
		fmt.Fprintf(w, "  digest:       %s\n", r.Digest)
	}
	if r.Source != "" {
		fmt.Fprintf(w, "  source:       %s\n", r.Source)
	}
	if r.Architecture != "" || r.OS != "" {
		fmt.Fprintf(w, "  platform:     %s/%s\n", r.OS, r.Architecture)
	}
	if r.Created != nil && !r.Created.IsZero() {
		fmt.Fprintf(w, "  created:      %s\n", r.Created.UTC().Format(time.RFC3339))
	}

	if r.KernelMin != "" {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  kernel min:   %s\n", r.KernelMin)
	}

	if hasOCIMetadata(r) {
		fmt.Fprintln(w)
		if r.Version != "" {
			fmt.Fprintf(w, "  version:      %s\n", r.Version)
		}
		if r.Revision != "" {
			fmt.Fprintf(w, "  revision:     %s\n", r.Revision)
		}
		if r.SourceURL != "" {
			fmt.Fprintf(w, "  source url:   %s\n", r.SourceURL)
		}
		if r.Documentation != "" {
			fmt.Fprintf(w, "  docs:         %s\n", r.Documentation)
		}
		if r.Authors != "" {
			fmt.Fprintf(w, "  authors:      %s\n", r.Authors)
		}
		if r.Licenses != "" {
			fmt.Fprintf(w, "  licenses:     %s\n", r.Licenses)
		}
		if r.Vendor != "" {
			fmt.Fprintf(w, "  vendor:       %s\n", r.Vendor)
		}
	}
}

func hasOCIMetadata(r Report) bool {
	return r.Version != "" || r.Revision != "" || r.SourceURL != "" ||
		r.Documentation != "" || r.Authors != "" || r.Licenses != "" || r.Vendor != ""
}

// Package status builds and renders the report consumed by `schedctl status`.
//
// The pure report-building logic lives here so it can be tested without a
// container runtime or kernel attached. The cmd layer is responsible for
// gathering inputs (managed containers, sched_ext state) and rendering.
package status

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"schedctl/internal/containers"
	"schedctl/internal/sched_ext"
)

// SchemaVersion is bumped on any breaking change to the JSON schema.
const SchemaVersion = "1"

// Stable status codes emitted in the JSON schema and in the text summary.
// Once published these strings are part of the documented schema and should
// not be renamed without a SchemaVersion bump.
const (
	StatusIdle            = "idle"
	StatusRunning         = "running"
	StatusOrphanedKernel  = "orphaned-kernel-state"
	StatusManagedDetached = "managed-detached"
	StatusManagedMismatch = "managed-mismatch"
	StatusMultipleManaged = "multiple-managed"
)

// Scheduler describes the scheduler that schedctl believes it is managing.
// Field tags define the wire format consumed by sked and other scripts.
type Scheduler struct {
	Name        string    `json:"name"`
	ContainerID string    `json:"container_id"`
	Image       string    `json:"image"`
	ImageID     string    `json:"image_id,omitempty"`
	PID         int       `json:"pid"`
	StartedAt   time.Time `json:"started_at"`
}

// Report is the top-level JSON document emitted by `schedctl status`.
type Report struct {
	SchemaVersion string          `json:"schema_version"`
	Status        string          `json:"status"`
	Driver        string          `json:"driver"`
	Scheduler     *Scheduler      `json:"scheduler,omitempty"`
	Kernel        sched_ext.State `json:"kernel"`
	Discrepancy   string          `json:"discrepancy,omitempty"`
}

// IsDiscrepancy reports whether the status indicates schedctl's view and the
// kernel's view do not agree, and the caller should exit non-zero.
func (r Report) IsDiscrepancy() bool {
	switch r.Status {
	case StatusOrphanedKernel, StatusManagedDetached, StatusManagedMismatch, StatusMultipleManaged:
		return true
	}
	return false
}

// Build applies the discrepancy rules. Pure function so it can be unit
// tested without a runtime or kernel attached.
func Build(driver string, managed []containers.Container, kernel sched_ext.State) Report {
	report := Report{
		SchemaVersion: SchemaVersion,
		Driver:        driver,
		Kernel:        kernel,
	}

	switch len(managed) {
	case 0:
		if kernel.Enabled || kernel.Ops != "" {
			report.Status = StatusOrphanedKernel
			report.Discrepancy = fmt.Sprintf(
				"kernel reports sched_ext ops %q attached but schedctl is not managing any scheduler",
				kernel.Ops,
			)
		} else {
			report.Status = StatusIdle
		}
	case 1:
		c := managed[0]
		report.Scheduler = &Scheduler{
			Name:        c.Name,
			ContainerID: c.ID,
			Image:       c.Image,
			ImageID:     c.ImageID,
			PID:         c.PID,
			StartedAt:   c.StartedAt.UTC(),
		}

		switch {
		case !kernel.Supported:
			report.Status = StatusManagedDetached
			report.Discrepancy = fmt.Sprintf(
				"schedctl is managing %q but the kernel does not support sched_ext",
				c.Name,
			)
		case !kernel.Enabled || kernel.Ops == "":
			report.Status = StatusManagedDetached
			report.Discrepancy = fmt.Sprintf(
				"schedctl is managing %q but the kernel reports no scheduler attached",
				c.Name,
			)
		case !matchesOps(c.Name, kernel.Ops):
			report.Status = StatusManagedMismatch
			report.Discrepancy = fmt.Sprintf(
				"schedctl is managing %q but the kernel has %q attached",
				c.Name, kernel.Ops,
			)
		default:
			report.Status = StatusRunning
		}
	default:
		names := make([]string, 0, len(managed))
		for _, c := range managed {
			names = append(names, c.Name)
		}
		report.Status = StatusMultipleManaged
		report.Discrepancy = fmt.Sprintf(
			"schedctl is managing %d schedulers (%s); only one BPF scheduler can be attached at a time",
			len(managed), strings.Join(names, ", "),
		)
	}

	return report
}

// matchesOps reports whether the scheduler name aligns with the ops name the
// kernel exposes. The match is lenient on purpose:
//
//   - scx schedulers conventionally prefix their binary with "scx_" while the
//     ops name does not (scx_rusty / rusty), so we strip that.
//   - some schedulers append metadata to the struct_ops name. scx_rusty for
//     example registers as e.g. "rusty_1.1.0_x86_64_unknown_linux_gnu".
//   - schedctl itself appends a random hex suffix to the container name on a
//     name collision (scx_rusty-a1b2c3).
//
// We treat names as matching when either the normalized container name or
// the ops name is a prefix of the other, bounded by "_" or "-".
func matchesOps(name, ops string) bool {
	if ops == "" {
		return false
	}
	norm := strings.TrimPrefix(strings.ToLower(name), "scx_")
	opsLower := strings.ToLower(ops)

	if norm == opsLower {
		return true
	}
	if hasTokenPrefix(opsLower, norm) {
		return true
	}
	if hasTokenPrefix(norm, opsLower) {
		return true
	}
	return false
}

// hasTokenPrefix reports whether s starts with prefix and the next character
// is a token boundary ("_" or "-").
func hasTokenPrefix(s, prefix string) bool {
	if prefix == "" || !strings.HasPrefix(s, prefix) {
		return false
	}
	if len(s) == len(prefix) {
		return true
	}
	switch s[len(prefix)] {
	case '_', '-':
		return true
	}
	return false
}

// WriteJSON encodes the report to w as indented JSON terminated by a newline.
func WriteJSON(w io.Writer, r Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// WriteText writes a human-readable summary of the report to w.
func WriteText(w io.Writer, r Report) {
	switch r.Status {
	case StatusIdle:
		fmt.Fprintln(w, "no scheduler running")
		fmt.Fprintf(w, "  driver: %s\n", r.Driver)
		fmt.Fprintf(w, "  kernel: %s\n", kernelLine(r.Kernel))
		return
	case StatusOrphanedKernel:
		fmt.Fprintln(w, "discrepancy: orphaned kernel state")
		fmt.Fprintf(w, "  driver: %s\n", r.Driver)
		fmt.Fprintf(w, "  kernel: %s\n", kernelLine(r.Kernel))
		fmt.Fprintf(w, "  detail: %s\n", r.Discrepancy)
		return
	case StatusMultipleManaged:
		fmt.Fprintln(w, "discrepancy: multiple managed schedulers")
		fmt.Fprintf(w, "  driver: %s\n", r.Driver)
		fmt.Fprintf(w, "  kernel: %s\n", kernelLine(r.Kernel))
		fmt.Fprintf(w, "  detail: %s\n", r.Discrepancy)
		return
	}

	switch r.Status {
	case StatusRunning:
		fmt.Fprintf(w, "scheduler running: %s\n", r.Scheduler.Name)
	case StatusManagedDetached:
		fmt.Fprintf(w, "discrepancy: %s container running but kernel detached\n", r.Scheduler.Name)
	case StatusManagedMismatch:
		fmt.Fprintf(w, "discrepancy: %s container running, kernel ops mismatch\n", r.Scheduler.Name)
	}

	fmt.Fprintf(w, "  driver:     %s\n", r.Driver)
	fmt.Fprintf(w, "  container:  %s\n", r.Scheduler.ContainerID)
	fmt.Fprintf(w, "  image:      %s\n", r.Scheduler.Image)
	if r.Scheduler.ImageID != "" {
		fmt.Fprintf(w, "  digest:     %s\n", r.Scheduler.ImageID)
	}
	fmt.Fprintf(w, "  pid:        %d\n", r.Scheduler.PID)
	if !r.Scheduler.StartedAt.IsZero() {
		fmt.Fprintf(w, "  started:    %s\n", r.Scheduler.StartedAt.Format(time.RFC3339))
	}
	fmt.Fprintf(w, "  kernel:     %s\n", kernelLine(r.Kernel))
	if r.Discrepancy != "" {
		fmt.Fprintf(w, "  detail:     %s\n", r.Discrepancy)
	}
}

func kernelLine(k sched_ext.State) string {
	if !k.Supported {
		return "sched_ext not supported"
	}
	state := "disabled"
	if k.Enabled {
		state = "enabled"
	}
	if k.Ops == "" {
		return state + ", no ops attached"
	}
	return fmt.Sprintf("%s, ops=%s", state, k.Ops)
}

package status_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"schedctl/internal/containers"
	"schedctl/internal/sched_ext"
	"schedctl/internal/status"
)

func TestBuildIdle(t *testing.T) {
	r := status.Build("podman", nil, sched_ext.State{Supported: true})

	assert.Equal(t, status.StatusIdle, r.Status)
	assert.Nil(t, r.Scheduler)
	assert.Empty(t, r.Discrepancy)
	assert.False(t, r.IsDiscrepancy())
}

func TestBuildRunning(t *testing.T) {
	started := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	managed := []containers.Container{{
		Name:      "scx_rusty",
		ID:        "abc123",
		PID:       4242,
		Image:     "ghcr.io/sched-ext/scx_rusty:latest",
		ImageID:   "sha256:cafebabe",
		StartedAt: started,
	}}
	kernel := sched_ext.State{Supported: true, Enabled: true, Ops: "rusty"}

	r := status.Build("podman", managed, kernel)

	assert.Equal(t, status.StatusRunning, r.Status)
	assert.Empty(t, r.Discrepancy)
	assert.False(t, r.IsDiscrepancy())
	if assert.NotNil(t, r.Scheduler) {
		assert.Equal(t, "scx_rusty", r.Scheduler.Name)
		assert.Equal(t, "abc123", r.Scheduler.ContainerID)
		assert.Equal(t, "ghcr.io/sched-ext/scx_rusty:latest", r.Scheduler.Image)
		assert.Equal(t, "sha256:cafebabe", r.Scheduler.ImageID)
		assert.Equal(t, 4242, r.Scheduler.PID)
		assert.Equal(t, started, r.Scheduler.StartedAt)
	}
}

func TestBuildRunningExactNameMatch(t *testing.T) {
	managed := []containers.Container{{Name: "rusty"}}
	kernel := sched_ext.State{Supported: true, Enabled: true, Ops: "rusty"}

	r := status.Build("podman", managed, kernel)

	assert.Equal(t, status.StatusRunning, r.Status)
}

func TestBuildOrphanedKernel(t *testing.T) {
	kernel := sched_ext.State{Supported: true, Enabled: true, Ops: "rusty"}

	r := status.Build("podman", nil, kernel)

	assert.Equal(t, status.StatusOrphanedKernel, r.Status)
	assert.True(t, r.IsDiscrepancy())
	assert.Contains(t, r.Discrepancy, "rusty")
}

func TestBuildManagedDetached(t *testing.T) {
	managed := []containers.Container{{Name: "scx_rusty"}}
	kernel := sched_ext.State{Supported: true, Enabled: false, Ops: ""}

	r := status.Build("podman", managed, kernel)

	assert.Equal(t, status.StatusManagedDetached, r.Status)
	assert.True(t, r.IsDiscrepancy())
}

// Kernel reports a non-empty ops name but state file says disabled (or is
// missing, leaving Enabled=false). The container name aligning with ops is
// not enough to call this "running" — sched_ext itself is not enabled.
func TestBuildManagedDetachedDisabledWithOps(t *testing.T) {
	managed := []containers.Container{{Name: "scx_rusty"}}
	kernel := sched_ext.State{Supported: true, Enabled: false, Ops: "rusty"}

	r := status.Build("podman", managed, kernel)

	assert.Equal(t, status.StatusManagedDetached, r.Status)
	assert.True(t, r.IsDiscrepancy())
	assert.Contains(t, r.Discrepancy, "scx_rusty")
}

// Kernel reports enabled but no ops name. No scheduler is actually attached.
func TestBuildManagedDetachedEnabledWithoutOps(t *testing.T) {
	managed := []containers.Container{{Name: "scx_rusty"}}
	kernel := sched_ext.State{Supported: true, Enabled: true, Ops: ""}

	r := status.Build("podman", managed, kernel)

	assert.Equal(t, status.StatusManagedDetached, r.Status)
	assert.True(t, r.IsDiscrepancy())
}

// Kernel does not support sched_ext at all but a scheduler container is
// running. Distinct detail message so the user is not told the kernel is
// "detached" when in fact sched_ext is missing entirely.
func TestBuildManagedDetachedKernelUnsupported(t *testing.T) {
	managed := []containers.Container{{Name: "scx_rusty"}}
	kernel := sched_ext.State{Supported: false}

	r := status.Build("podman", managed, kernel)

	assert.Equal(t, status.StatusManagedDetached, r.Status)
	assert.True(t, r.IsDiscrepancy())
	assert.Contains(t, r.Discrepancy, "does not support sched_ext")
}

func TestBuildManagedMismatch(t *testing.T) {
	managed := []containers.Container{{Name: "scx_rusty"}}
	kernel := sched_ext.State{Supported: true, Enabled: true, Ops: "lavd"}

	r := status.Build("podman", managed, kernel)

	assert.Equal(t, status.StatusManagedMismatch, r.Status)
	assert.True(t, r.IsDiscrepancy())
	assert.Contains(t, r.Discrepancy, "scx_rusty")
	assert.Contains(t, r.Discrepancy, "lavd")
}

// scx_rusty registers a struct_ops name that embeds the build version and
// target triple, e.g. "rusty_1.1.0_x86_64_unknown_linux_gnu". The match must
// not flag that as a discrepancy.
func TestBuildRunningWithVersionedOps(t *testing.T) {
	managed := []containers.Container{{Name: "scx_rusty"}}
	kernel := sched_ext.State{
		Supported: true,
		Enabled:   true,
		Ops:       "rusty_1.1.0_x86_64_unknown_linux_gnu",
	}

	r := status.Build("podman", managed, kernel)

	assert.Equal(t, status.StatusRunning, r.Status)
}

// On a name collision podman.Run appends a 6-char hex suffix, e.g.
// "scx_rusty-a1b2c3". That must still be considered a match against ops
// "rusty".
func TestBuildRunningWithRandomSuffixedContainer(t *testing.T) {
	managed := []containers.Container{{Name: "scx_rusty-a1b2c3"}}
	kernel := sched_ext.State{Supported: true, Enabled: true, Ops: "rusty"}

	r := status.Build("podman", managed, kernel)

	assert.Equal(t, status.StatusRunning, r.Status)
}

func TestBuildMultipleManaged(t *testing.T) {
	managed := []containers.Container{
		{Name: "scx_rusty"},
		{Name: "scx_lavd"},
	}
	kernel := sched_ext.State{Supported: true, Enabled: true, Ops: "rusty"}

	r := status.Build("podman", managed, kernel)

	assert.Equal(t, status.StatusMultipleManaged, r.Status)
	assert.True(t, r.IsDiscrepancy())
	assert.Contains(t, r.Discrepancy, "scx_rusty")
	assert.Contains(t, r.Discrepancy, "scx_lavd")
}

// TestJSONSchemaStability locks the documented field names so downstream
// consumers (sked, user scripts) can rely on them.
func TestJSONSchemaStability(t *testing.T) {
	managed := []containers.Container{{
		Name:      "scx_rusty",
		ID:        "abc",
		PID:       1,
		Image:     "img",
		ImageID:   "sha256:deadbeef",
		StartedAt: time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
	}}
	kernel := sched_ext.State{Supported: true, Enabled: true, Ops: "rusty"}

	r := status.Build("podman", managed, kernel)

	var buf bytes.Buffer
	if err := status.WriteJSON(&buf, r); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, key := range []string{"schema_version", "status", "driver", "scheduler", "kernel"} {
		_, ok := raw[key]
		assert.True(t, ok, "top-level key %q must be present", key)
	}

	assert.Equal(t, status.SchemaVersion, raw["schema_version"])

	sched, ok := raw["scheduler"].(map[string]any)
	if assert.True(t, ok, "scheduler must serialize as a JSON object") {
		for _, key := range []string{"name", "container_id", "image", "image_id", "pid", "started_at"} {
			_, present := sched[key]
			assert.True(t, present, "scheduler.%s must be present", key)
		}
	}

	kernelRaw, ok := raw["kernel"].(map[string]any)
	if assert.True(t, ok, "kernel must serialize as a JSON object") {
		for _, key := range []string{"supported", "enabled", "ops"} {
			_, present := kernelRaw[key]
			assert.True(t, present, "kernel.%s must be present", key)
		}
	}
}

func TestJSONOmitsSchedulerWhenIdle(t *testing.T) {
	r := status.Build("podman", nil, sched_ext.State{Supported: true})

	var buf bytes.Buffer
	if err := status.WriteJSON(&buf, r); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	_, hasScheduler := raw["scheduler"]
	assert.False(t, hasScheduler, "scheduler key must be omitted when no scheduler is running")

	_, hasDiscrepancy := raw["discrepancy"]
	assert.False(t, hasDiscrepancy, "discrepancy key must be omitted when there is none")
}

package doctor_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"schedctl/internal/doctor"
)

func passing(id string, sev doctor.Severity) doctor.Check {
	return doctor.Check{
		ID:          id,
		Description: "passing " + id,
		Severity:    sev,
		Remediation: "n/a",
		Func:        func() (doctor.Status, string) { return doctor.StatusPass, "ok" },
	}
}

func failing(id string, sev doctor.Severity) doctor.Check {
	return doctor.Check{
		ID:          id,
		Description: "failing " + id,
		Severity:    sev,
		Remediation: "fix " + id,
		Func:        func() (doctor.Status, string) { return doctor.StatusFail, "broken" },
	}
}

func skipping(id string, sev doctor.Severity) doctor.Check {
	return doctor.Check{
		ID:          id,
		Description: "skipping " + id,
		Severity:    sev,
		Remediation: "no remedy",
		Func:        func() (doctor.Status, string) { return doctor.StatusSkip, "indeterminate" },
	}
}

func TestRunTalliesResults(t *testing.T) {
	report := doctor.Run([]doctor.Check{
		passing("a", doctor.SeverityError),
		failing("b", doctor.SeverityWarn),
		skipping("c", doctor.SeverityError),
	})

	assert.Equal(t, 1, report.Summary.Passed)
	assert.Equal(t, 1, report.Summary.Failed)
	assert.Equal(t, 1, report.Summary.Skipped)
	assert.Equal(t, 0, report.Summary.Blocking, "warn-level fail must not block")
	assert.False(t, report.HasBlockingFailures())
}

func TestRunFlagsBlockingOnErrorFailures(t *testing.T) {
	report := doctor.Run([]doctor.Check{
		failing("a", doctor.SeverityError),
		failing("b", doctor.SeverityWarn),
	})

	assert.Equal(t, 2, report.Summary.Failed)
	assert.Equal(t, 1, report.Summary.Blocking)
	assert.True(t, report.HasBlockingFailures())
}

func TestRunPreservesCheckMetadata(t *testing.T) {
	report := doctor.Run([]doctor.Check{failing("kernel.x", doctor.SeverityError)})

	assert.Len(t, report.Checks, 1)
	got := report.Checks[0]
	assert.Equal(t, "kernel.x", got.ID)
	assert.Equal(t, doctor.SeverityError, got.Severity)
	assert.Equal(t, doctor.StatusFail, got.Status)
	assert.Equal(t, "broken", got.Detail)
	assert.Equal(t, "fix kernel.x", got.Remediation)
}

func TestWriteJSONShape(t *testing.T) {
	report := doctor.Run([]doctor.Check{
		passing("a", doctor.SeverityError),
		failing("b", doctor.SeverityError),
	})

	var buf bytes.Buffer
	assert.NoError(t, doctor.WriteJSON(&buf, report))

	var parsed struct {
		Checks []struct {
			ID          string `json:"id"`
			Severity    string `json:"severity"`
			Status      string `json:"status"`
			Remediation string `json:"remediation"`
		} `json:"checks"`
		Summary struct {
			Passed   int `json:"passed"`
			Failed   int `json:"failed"`
			Skipped  int `json:"skipped"`
			Blocking int `json:"blocking"`
		} `json:"summary"`
	}
	assert.NoError(t, json.Unmarshal(buf.Bytes(), &parsed))

	assert.Len(t, parsed.Checks, 2)
	assert.Equal(t, "a", parsed.Checks[0].ID)
	assert.Equal(t, "pass", parsed.Checks[0].Status)
	assert.Equal(t, "fail", parsed.Checks[1].Status)
	assert.Equal(t, "error", parsed.Checks[1].Severity)
	assert.NotEmpty(t, parsed.Checks[1].Remediation)
	assert.Equal(t, 1, parsed.Summary.Passed)
	assert.Equal(t, 1, parsed.Summary.Failed)
	assert.Equal(t, 1, parsed.Summary.Blocking)
}

func TestWriteJSONEmptyReportIsNotNull(t *testing.T) {
	var buf bytes.Buffer
	assert.NoError(t, doctor.WriteJSON(&buf, doctor.Run(nil)))
	assert.Contains(t, buf.String(), `"checks": []`)
}

func TestWriteTextHumanReadable(t *testing.T) {
	report := doctor.Run([]doctor.Check{
		passing("kernel.version", doctor.SeverityError),
		failing("caps.cap_bpf", doctor.SeverityError),
		skipping("kernel.config", doctor.SeverityError),
	})

	var buf bytes.Buffer
	assert.NoError(t, doctor.WriteText(&buf, report))

	out := buf.String()
	assert.Contains(t, out, "[PASS]")
	assert.Contains(t, out, "[ERROR]", "failed error-severity check should render as [ERROR]")
	assert.Contains(t, out, "[SKIP]")
	assert.Contains(t, out, "kernel.version")
	assert.Contains(t, out, "caps.cap_bpf")
	assert.Contains(t, out, "Remediation:")
	assert.Contains(t, out, "fix caps.cap_bpf")
	assert.Contains(t, out, "1 passed, 1 failed, 1 skipped (1 blocking)")
}

func TestDefaultChecksHaveStableIDsAndRemediation(t *testing.T) {
	checks := doctor.DefaultChecks()
	assert.NotEmpty(t, checks)

	wantIDs := []string{
		"kernel.version",
		"kernel.sched_ext",
		"kernel.btf",
		"kernel.config",
		"caps.cap_bpf",
		"caps.cap_sys_admin",
		"caps.cap_perfmon",
		"runtime.podman_socket",
		"runtime.containerd_socket",
		"runtime.any",
	}

	got := make(map[string]doctor.Check, len(checks))
	for _, c := range checks {
		got[c.ID] = c
	}
	for _, id := range wantIDs {
		c, ok := got[id]
		assert.True(t, ok, "expected default check %q", id)
		if !ok {
			continue
		}
		assert.NotEmpty(t, c.Description, "%s: description must not be empty", id)
		assert.NotEmpty(t, c.Remediation, "%s: remediation hint must not be empty", id)
		assert.Contains(t, []doctor.Severity{doctor.SeverityInfo, doctor.SeverityWarn, doctor.SeverityError}, c.Severity,
			"%s: severity must be one of info/warn/error", id)
		assert.False(t, strings.Contains(c.Remediation, "\n"), "%s: remediation must be one line", id)
	}
}

// Package doctor implements host readiness checks for sched_ext schedulers.
//
// Each Check has a stable ID, a severity, and a remediation hint. Run executes
// a list of checks and produces a Report consumable by both humans (via
// WriteText) and tooling (via WriteJSON).
package doctor

type Severity string

const (
	SeverityInfo  Severity = "info"
	SeverityWarn  Severity = "warn"
	SeverityError Severity = "error"
)

type Status string

const (
	StatusPass Status = "pass"
	StatusFail Status = "fail"
	StatusSkip Status = "skip"
)

type CheckFunc func() (Status, string)

type Check struct {
	ID          string
	Description string
	Severity    Severity
	Remediation string
	Func        CheckFunc
}

type Result struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
	Status      Status   `json:"status"`
	Detail      string   `json:"detail,omitempty"`
	Remediation string   `json:"remediation"`
}

type Summary struct {
	Passed   int `json:"passed"`
	Failed   int `json:"failed"`
	Skipped  int `json:"skipped"`
	Blocking int `json:"blocking"`
}

type Report struct {
	Checks  []Result `json:"checks"`
	Summary Summary  `json:"summary"`
}

func (r Report) HasBlockingFailures() bool {
	return r.Summary.Blocking > 0
}

func Run(checks []Check) Report {
	results := make([]Result, 0, len(checks))
	summary := Summary{}
	for _, c := range checks {
		status, detail := c.Func()
		results = append(results, Result{
			ID:          c.ID,
			Description: c.Description,
			Severity:    c.Severity,
			Status:      status,
			Detail:      detail,
			Remediation: c.Remediation,
		})
		switch status {
		case StatusPass:
			summary.Passed++
		case StatusFail:
			summary.Failed++
			if c.Severity == SeverityError {
				summary.Blocking++
			}
		case StatusSkip:
			summary.Skipped++
		}
	}
	return Report{Checks: results, Summary: summary}
}

func DefaultChecks() []Check {
	kernel, caps, runtime := kernelChecks(), capsChecks(), runtimeChecks()
	checks := make([]Check, 0, len(kernel)+len(caps)+len(runtime))
	checks = append(checks, kernel...)
	checks = append(checks, caps...)
	checks = append(checks, runtime...)
	return checks
}

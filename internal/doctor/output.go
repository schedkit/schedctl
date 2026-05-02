package doctor

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

func WriteJSON(w io.Writer, r Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

func WriteText(w io.Writer, r Report) error {
	if _, err := fmt.Fprintln(w, "schedctl doctor — host readiness check"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, c := range r.Checks {
		label := statusLabel(c.Status, c.Severity)
		detail := c.Detail
		if detail == "" {
			detail = c.Description
		}
		if _, err := fmt.Fprintf(tw, "[%s]\t%s\t%s\n", label, c.ID, detail); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	remediations := make([]Result, 0)
	for _, c := range r.Checks {
		if c.Status == StatusFail || c.Status == StatusSkip {
			remediations = append(remediations, c)
		}
	}
	if len(remediations) > 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "Remediation:"); err != nil {
			return err
		}
		for _, c := range remediations {
			if _, err := fmt.Fprintf(w, "  - %s: %s\n", c.ID, c.Remediation); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "Summary: %d passed, %d failed, %d skipped (%d blocking)\n",
		r.Summary.Passed, r.Summary.Failed, r.Summary.Skipped, r.Summary.Blocking)
	return err
}

func statusLabel(s Status, sev Severity) string {
	switch s {
	case StatusPass:
		return "PASS"
	case StatusSkip:
		return "SKIP"
	case StatusFail:
		return strings.ToUpper(string(sev))
	default:
		return strings.ToUpper(string(s))
	}
}

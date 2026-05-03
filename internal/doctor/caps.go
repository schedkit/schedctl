package doctor

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	capSysAdmin uint = 21
	capPerfmon  uint = 38
	capBPF      uint = 39
)

func capsChecks() []Check {
	return []Check{
		{
			ID:          "caps.cap_bpf",
			Description: "process holds CAP_BPF in its effective set",
			Severity:    SeverityError,
			Remediation: "run schedctl as root or grant CAP_BPF to the binary (setcap cap_bpf+ep)",
			Func:        capCheckFunc(capBPF),
		},
		{
			ID:          "caps.cap_sys_admin",
			Description: "process holds CAP_SYS_ADMIN in its effective set",
			Severity:    SeverityError,
			Remediation: "run schedctl as root or grant CAP_SYS_ADMIN to the binary (setcap cap_sys_admin+ep)",
			Func:        capCheckFunc(capSysAdmin),
		},
		{
			ID:          "caps.cap_perfmon",
			Description: "process holds CAP_PERFMON in its effective set",
			Severity:    SeverityError,
			Remediation: "run schedctl as root or grant CAP_PERFMON to the binary (setcap cap_perfmon+ep)",
			Func:        capCheckFunc(capPerfmon),
		},
	}
}

func capCheckFunc(bit uint) CheckFunc {
	return func() (Status, string) {
		eff, err := readEffectiveCaps()
		if err != nil {
			return StatusFail, err.Error()
		}
		if eff&(uint64(1)<<bit) != 0 {
			return StatusPass, fmt.Sprintf("CapEff bit %d set", bit)
		}
		return StatusFail, fmt.Sprintf("CapEff bit %d not set (CapEff=0x%016x)", bit, eff)
	}
}

func readEffectiveCaps() (uint64, error) {
	f, err := os.Open("/proc/self/status")
	if err != nil {
		return 0, fmt.Errorf("cannot read /proc/self/status: %w", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "CapEff:") {
			continue
		}
		raw := strings.TrimSpace(strings.TrimPrefix(line, "CapEff:"))
		parsed, err := strconv.ParseUint(raw, 16, 64)
		if err != nil {
			return 0, fmt.Errorf("malformed CapEff line: %q", line)
		}
		return parsed, nil
	}
	return 0, fmt.Errorf("CapEff not found in /proc/self/status")
}

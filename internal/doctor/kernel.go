package doctor

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	minKernelMajor = 6
	minKernelMinor = 12
)

func requiredKernelConfigs() []string {
	return []string{
		"CONFIG_BPF",
		"CONFIG_BPF_SYSCALL",
		"CONFIG_BPF_JIT",
		"CONFIG_DEBUG_INFO_BTF",
		"CONFIG_SCHED_CLASS_EXT",
	}
}

func kernelChecks() []Check {
	return []Check{
		{
			ID:          "kernel.version",
			Description: "Linux kernel version supports sched_ext",
			Severity:    SeverityError,
			Remediation: fmt.Sprintf("upgrade to a kernel >= %d.%d which mainlines sched_ext", minKernelMajor, minKernelMinor),
			Func:        checkKernelVersion,
		},
		{
			ID:          "kernel.sched_ext",
			Description: "sched_ext interface is exposed under /sys/kernel/sched_ext",
			Severity:    SeverityError,
			Remediation: "boot a kernel built with CONFIG_SCHED_CLASS_EXT=y",
			Func:        checkSchedExt,
		},
		{
			ID:          "kernel.btf",
			Description: "vmlinux BTF is available for CO-RE BPF programs",
			Severity:    SeverityError,
			Remediation: "build the running kernel with CONFIG_DEBUG_INFO_BTF=y",
			Func:        checkBTF,
		},
		{
			ID:          "kernel.config",
			Description: "Required kernel CONFIG_* flags are enabled",
			Severity:    SeverityError,
			Remediation: "rebuild the kernel with CONFIG_BPF, CONFIG_BPF_SYSCALL, CONFIG_BPF_JIT," +
				" CONFIG_DEBUG_INFO_BTF and CONFIG_SCHED_CLASS_EXT",
			Func: checkKernelConfig,
		},
	}
}

func unameRelease() (string, error) {
	var u unix.Utsname
	if err := unix.Uname(&u); err != nil {
		return "", err
	}
	end := 0
	for end < len(u.Release) && u.Release[end] != 0 {
		end++
	}
	return string(u.Release[:end]), nil
}

func parseKernelVersion(s string) (int, int, error) {
	parts := strings.SplitN(s, ".", 3)
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("unrecognized kernel version %q", s)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("unrecognized kernel version %q", s)
	}
	minor := parts[1]
	end := 0
	for end < len(minor) && minor[end] >= '0' && minor[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, 0, fmt.Errorf("unrecognized kernel version %q", s)
	}
	minorN, err := strconv.Atoi(minor[:end])
	if err != nil {
		return 0, 0, fmt.Errorf("unrecognized kernel version %q", s)
	}
	return major, minorN, nil
}

func checkKernelVersion() (Status, string) {
	rel, err := unameRelease()
	if err != nil {
		return StatusSkip, fmt.Sprintf("uname failed: %v", err)
	}
	major, minor, err := parseKernelVersion(rel)
	if err != nil {
		return StatusSkip, err.Error()
	}
	if major > minKernelMajor || (major == minKernelMajor && minor >= minKernelMinor) {
		return StatusPass, rel
	}
	return StatusFail, fmt.Sprintf("running %s, need >= %d.%d", rel, minKernelMajor, minKernelMinor)
}

func checkSchedExt() (Status, string) {
	if info, err := os.Stat("/sys/kernel/sched_ext"); err == nil && info.IsDir() {
		return StatusPass, "/sys/kernel/sched_ext is present"
	}
	return StatusFail, "/sys/kernel/sched_ext is not present"
}

func checkBTF() (Status, string) {
	info, err := os.Stat("/sys/kernel/btf/vmlinux")
	if err != nil {
		return StatusFail, "/sys/kernel/btf/vmlinux is not available"
	}
	if info.Size() == 0 {
		return StatusFail, "/sys/kernel/btf/vmlinux is empty"
	}
	return StatusPass, fmt.Sprintf("/sys/kernel/btf/vmlinux (%d bytes)", info.Size())
}

func checkKernelConfig() (Status, string) {
	cfg, err := readKernelConfig()
	if err != nil {
		return StatusFail, "kernel config not readable; install kernel-headers or enable CONFIG_IKCONFIG_PROC"
	}
	required := requiredKernelConfigs()
	missing := make([]string, 0, len(required))
	for _, key := range required {
		v, ok := cfg[key]
		if !ok || (v != "y" && v != "m") {
			missing = append(missing, key)
		}
	}
	if len(missing) == 0 {
		return StatusPass, "all required CONFIG_* flags enabled"
	}
	return StatusFail, "missing or disabled: " + strings.Join(missing, ", ")
}

func readKernelConfig() (map[string]string, error) {
	rel, _ := unameRelease()
	candidates := []string{
		"/proc/config.gz",
		"/boot/config-" + rel,
		"/lib/modules/" + rel + "/config",
	}
	for _, path := range candidates {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		var reader io.Reader = f
		if strings.HasSuffix(path, ".gz") {
			gz, err := gzip.NewReader(f)
			if err != nil {
				_ = f.Close()
				continue
			}
			defer gz.Close()
			reader = gz
		}
		out := make(map[string]string)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			eq := strings.IndexByte(line, '=')
			if eq <= 0 {
				continue
			}
			out[line[:eq]] = line[eq+1:]
		}
		if err := scanner.Err(); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		_ = f.Close()
		return out, nil
	}
	return nil, fmt.Errorf("kernel config not found")
}

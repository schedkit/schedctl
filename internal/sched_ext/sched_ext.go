// Package sched_ext reads kernel sched_ext state from sysfs.
//
// /sys/kernel/sched_ext is exposed by kernels with sched_ext support and
// reflects the currently attached BPF scheduler, if any.
package sched_ext

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const DefaultSysfsRoot = "/sys/kernel/sched_ext"

// State captures what the kernel reports about sched_ext.
//
// Supported is false when /sys/kernel/sched_ext does not exist (kernel was
// built without sched_ext or it has not been initialized).
//
// Enabled reflects /sys/kernel/sched_ext/state. Ops is the name of the
// attached scheduler ops as reported by /sys/kernel/sched_ext/root/ops, or
// empty if no scheduler is attached.
type State struct {
	Supported bool   `json:"supported"`
	Enabled   bool   `json:"enabled"`
	Ops       string `json:"ops"`
}

// Read returns the current sched_ext state by reading the default sysfs
// location. Missing files are treated as "not attached" rather than errors.
func Read() (State, error) {
	return ReadFrom(DefaultSysfsRoot)
}

// ReadFrom reads sched_ext state from the given root path. Useful for tests.
func ReadFrom(root string) (State, error) {
	var st State

	if _, err := os.Stat(root); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return st, nil
		}
		return st, err
	}
	st.Supported = true

	if data, err := os.ReadFile(filepath.Join(root, "state")); err == nil {
		st.Enabled = strings.TrimSpace(string(data)) == "enabled"
	} else if !errors.Is(err, fs.ErrNotExist) {
		return st, err
	}

	if data, err := os.ReadFile(filepath.Join(root, "root", "ops")); err == nil {
		st.Ops = strings.TrimSpace(string(data))
	} else if !errors.Is(err, fs.ErrNotExist) {
		return st, err
	}

	return st, nil
}

package sched_ext_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"schedctl/internal/sched_ext"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestReadFromMissingRoot(t *testing.T) {
	dir := t.TempDir()
	st, err := sched_ext.ReadFrom(filepath.Join(dir, "does-not-exist"))

	assert.NoError(t, err)
	assert.False(t, st.Supported)
	assert.False(t, st.Enabled)
	assert.Empty(t, st.Ops)
}

func TestReadFromEnabledWithOps(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "state"), "enabled\n")
	writeFile(t, filepath.Join(dir, "root", "ops"), "rusty\n")

	st, err := sched_ext.ReadFrom(dir)

	assert.NoError(t, err)
	assert.True(t, st.Supported)
	assert.True(t, st.Enabled)
	assert.Equal(t, "rusty", st.Ops)
}

func TestReadFromSupportedButDisabled(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "state"), "disabled\n")

	st, err := sched_ext.ReadFrom(dir)

	assert.NoError(t, err)
	assert.True(t, st.Supported)
	assert.False(t, st.Enabled)
	assert.Empty(t, st.Ops)
}

func TestReadFromOpsWithoutState(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "root", "ops"), "lavd\n")

	st, err := sched_ext.ReadFrom(dir)

	assert.NoError(t, err)
	assert.True(t, st.Supported)
	assert.False(t, st.Enabled)
	assert.Equal(t, "lavd", st.Ops)
}

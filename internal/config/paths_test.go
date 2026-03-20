package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultConfigDir_TRAY_CONFIG_DIR(t *testing.T) {
	t.Setenv("TRAY_CONFIG_DIR", "/custom/tray")
	t.Setenv("XDG_CONFIG_HOME", "/xdg") // should be ignored
	got := DefaultConfigDir()
	require.Equal(t, "/custom/tray", got)
}

func TestDefaultConfigDir_XDG(t *testing.T) {
	t.Setenv("TRAY_CONFIG_DIR", "")
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	want := filepath.Join(dir, "tray")
	require.Equal(t, want, DefaultConfigDir())
}

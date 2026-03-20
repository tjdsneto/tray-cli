package config

import (
	"path/filepath"
	"runtime"
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

func TestDefaultConfigDir_UnixLike_homeConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	t.Setenv("TRAY_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	want := filepath.Join(home, ".config", "tray")
	require.Equal(t, want, DefaultConfigDir())
}

func TestDefaultConfigDir_Windows_APPDATA(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip()
	}
	t.Setenv("TRAY_CONFIG_DIR", "")
	appData := t.TempDir()
	t.Setenv("APPDATA", appData)
	t.Setenv("XDG_CONFIG_HOME", "") // irrelevant on Windows
	want := filepath.Join(appData, "tray")
	require.Equal(t, want, DefaultConfigDir())
}

package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// DefaultConfigDir returns the tray config directory (no I/O).
//
// Resolution order (OS-specific defaults are abstracted here so macOS, Linux, and Windows
// each get a conventional layout; callers should not hardcode paths):
//
//   1. TRAY_CONFIG_DIR — explicit override (any OS; use for tests or custom locations).
//   2. Windows: %APPDATA%\tray (or os.UserConfigDir()/tray if APPDATA is unset).
//   3. Unix-like (Linux, macOS, *BSD, etc.): $XDG_CONFIG_HOME/tray if XDG_CONFIG_HOME is set,
//      else ~/.config/tray (same as many CLIs; keeps Linux and macOS aligned).
//
// All paths use filepath.Join for correct separators on each platform.
func DefaultConfigDir() string {
	if d := os.Getenv("TRAY_CONFIG_DIR"); d != "" {
		return d
	}
	if runtime.GOOS == "windows" {
		return defaultConfigDirWindows()
	}
	return defaultConfigDirUnixLike()
}

func defaultConfigDirWindows() string {
	if d := os.Getenv("APPDATA"); d != "" {
		return filepath.Join(d, "tray")
	}
	if d, err := os.UserConfigDir(); err == nil {
		return filepath.Join(d, "tray")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "tray")
	}
	return filepath.Join(home, ".config", "tray")
}

func defaultConfigDirUnixLike() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "tray")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".config", "tray")
	}
	return filepath.Join(home, ".config", "tray")
}

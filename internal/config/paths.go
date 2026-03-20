package config

import (
	"os"
	"path/filepath"
)

// DefaultConfigDir returns the tray config directory (no I/O).
// Order: TRAY_CONFIG_DIR, then XDG_CONFIG_HOME/tray, then ~/.config/tray.
func DefaultConfigDir() string {
	if d := os.Getenv("TRAY_CONFIG_DIR"); d != "" {
		return d
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "tray")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".config", "tray")
	}
	return filepath.Join(home, ".config", "tray")
}

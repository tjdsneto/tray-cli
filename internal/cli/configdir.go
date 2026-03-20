package cli

import "github.com/tjdsneto/tray-cli/internal/config"

// ConfigDir returns --config-dir when set, else default tray config path.
func ConfigDir() string {
	if configDirFlag != "" {
		return configDirFlag
	}
	return config.DefaultConfigDir()
}

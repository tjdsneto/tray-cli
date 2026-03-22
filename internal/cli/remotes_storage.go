package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const remotesFileName = "remotes.json"

// remotesFile is persisted next to credentials (0600).
type remotesFile struct {
	Aliases map[string]string `json:"aliases"`
}

func loadRemotes(configDir string) (remotesFile, error) {
	p := filepath.Join(configDir, remotesFileName)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return remotesFile{Aliases: map[string]string{}}, nil
		}
		return remotesFile{}, err
	}
	var f remotesFile
	if err := json.Unmarshal(data, &f); err != nil {
		return remotesFile{}, err
	}
	if f.Aliases == nil {
		f.Aliases = map[string]string{}
	}
	return f, nil
}

func saveRemotes(configDir string, f remotesFile) error {
	if f.Aliases == nil {
		f.Aliases = map[string]string{}
	}
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	p := filepath.Join(configDir, remotesFileName)
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// loadRemoteAliases returns lowercased alias → tray id for resolveTrayRef.
func loadRemoteAliases(configDir string) map[string]string {
	f, err := loadRemotes(configDir)
	if err != nil {
		return map[string]string{}
	}
	m := make(map[string]string, len(f.Aliases))
	for k, v := range f.Aliases {
		ks := strings.ToLower(strings.TrimSpace(k))
		if ks == "" {
			continue
		}
		m[ks] = strings.TrimSpace(v)
	}
	return m
}

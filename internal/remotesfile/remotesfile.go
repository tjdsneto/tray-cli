// Package remotesfile persists tray remote aliases (remotes.json) next to credentials.
package remotesfile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const fileName = "remotes.json"

// File is the on-disk shape for remotes.json (0600).
type File struct {
	Aliases map[string]string `json:"aliases"`
}

// Load reads remotes.json or returns an empty file if missing.
func Load(configDir string) (File, error) {
	p := filepath.Join(configDir, fileName)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return File{Aliases: map[string]string{}}, nil
		}
		return File{}, err
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return File{}, err
	}
	if f.Aliases == nil {
		f.Aliases = map[string]string{}
	}
	return f, nil
}

// Save writes remotes.json atomically.
func Save(configDir string, f File) error {
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
	p := filepath.Join(configDir, fileName)
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// AliasesMap returns lowercased alias → tray id for tray resolution (best-effort empty map on read error).
func AliasesMap(configDir string) map[string]string {
	f, err := Load(configDir)
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

package localtray

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const indexFileName = "index.json"

// TrayRecord is one registered local tray.
type TrayRecord struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Name      string    `json:"name,omitempty"`
	Path      string    `json:"path,omitempty"`
	RepoRoot  string    `json:"repo_root,omitempty"`
	Branch    string    `json:"branch,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	ItemsFile string    `json:"items_file"`
}

// Index lists every local tray that was ever materialized.
type Index struct {
	Trays []TrayRecord `json:"trays"`
}

// Find returns a tray record by canonical id.
func (idx Index) Find(id string) (TrayRecord, bool) {
	id = strings.TrimSpace(id)
	for i := range idx.Trays {
		if idx.Trays[i].ID == id {
			return idx.Trays[i], true
		}
	}
	return TrayRecord{}, false
}

// Globals returns global tray records in index order.
func (idx Index) Globals() []TrayRecord {
	var out []TrayRecord
	for i := range idx.Trays {
		if idx.Trays[i].Kind == KindGlobal {
			out = append(out, idx.Trays[i])
		}
	}
	return out
}

// LoadIndex reads local/index.json or returns an empty index.
func LoadIndex(localDir string) (Index, error) {
	p := filepath.Join(localDir, indexFileName)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return Index{}, nil
		}
		return Index{}, err
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return Index{}, err
	}
	return idx, nil
}

// SaveIndex writes local/index.json atomically.
func SaveIndex(localDir string, idx Index) error {
	if err := os.MkdirAll(localDir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	p := filepath.Join(localDir, indexFileName)
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// LocalDir returns the subdirectory under configDir where local trays live.
func LocalDir(configDir string) string {
	return filepath.Join(configDir, "local")
}

// ItemsDir returns the directory for jsonl item files.
func ItemsDir(localDir string) string {
	return filepath.Join(localDir, "items")
}

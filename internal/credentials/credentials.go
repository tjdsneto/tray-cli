package credentials

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

const fileName = "credentials.json"

// File is persisted session data (0600).
type File struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	UserID       string `json:"user_id"`
}

// Path returns the credentials file path under configDir.
func Path(configDir string) string {
	return filepath.Join(configDir, fileName)
}

// Save writes credentials atomically (same dir, rename).
func Save(configDir string, f File) error {
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("credentials: mkdir: %w", err)
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("credentials: encode: %w", err)
	}
	p := Path(configDir)
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("credentials: write temp: %w", err)
	}
	if err := os.Rename(tmp, p); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("credentials: rename: %w", err)
	}
	return nil
}

// Load reads credentials from configDir.
func Load(configDir string) (File, error) {
	p := Path(configDir)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return File{}, fmt.Errorf("credentials: %w", ErrNotLoggedIn)
		}
		return File{}, fmt.Errorf("credentials: read: %w", err)
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return File{}, fmt.Errorf("credentials: decode: %w", err)
	}
	return f, nil
}

// ErrNotLoggedIn means no saved credentials.
var ErrNotLoggedIn = errors.New("not logged in; run tray login")

// Session maps stored credentials to a domain session.
func (f File) Session() domain.Session {
	return domain.Session{
		AccessToken: f.AccessToken,
		UserID:      f.UserID,
	}
}

package postgrest

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

// PostgREST may return one row as an object or as a single-element array.
func parseCreatedTray(raw []byte) (*trayRow, error) {
	var rows []trayRow
	if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 {
		return &rows[0], nil
	}
	var one trayRow
	if err := json.Unmarshal(raw, &one); err == nil && one.ID != "" {
		return &one, nil
	}
	return nil, fmt.Errorf("postgrest: parse create tray response: %s", strings.TrimSpace(string(raw)))
}

type trayRow struct {
	ID          string  `json:"id"`
	OwnerID     string  `json:"owner_id"`
	Name        string  `json:"name"`
	InviteToken *string `json:"invite_token"`
	CreatedAt   string  `json:"created_at"`
}

func trayFromRow(r trayRow) (*domain.Tray, error) {
	t, err := parseTime(r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("postgrest: tray created_at: %w", err)
	}
	return &domain.Tray{
		ID:          r.ID,
		OwnerID:     r.OwnerID,
		Name:        r.Name,
		InviteToken: r.InviteToken,
		CreatedAt:   t,
	}, nil
}

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
	}
	return t, err
}

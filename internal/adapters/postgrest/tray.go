package postgrest

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/timex"
)

type trayRow struct {
	ID          string           `json:"id"`
	OwnerID     string           `json:"owner_id"`
	Name        string           `json:"name"`
	InviteToken *string          `json:"invite_token"`
	CreatedAt   string           `json:"created_at"`
	Items       []trayItemsCount `json:"items,omitempty"`
}

// trayItemsCount is PostgREST embed shape: items(count) → [{"count": n}].
type trayItemsCount struct {
	Count int `json:"count"`
}

// joinTrayRequest is the JSON body for POST /rest/v1/rpc/join_tray.
type joinTrayRequest struct {
	PInviteToken string `json:"p_invite_token"`
}

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

func (r trayRow) ToDomain() (*domain.Tray, error) {
	t, err := timex.ParseRFC3339OrNano(r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("postgrest: tray created_at: %w", err)
	}
	n := 0
	if len(r.Items) > 0 {
		n = r.Items[0].Count
	}
	return &domain.Tray{
		ID:          r.ID,
		OwnerID:     r.OwnerID,
		Name:        r.Name,
		InviteToken: r.InviteToken,
		CreatedAt:   t,
		ItemCount:   n,
	}, nil
}

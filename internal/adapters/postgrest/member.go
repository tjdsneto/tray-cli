package postgrest

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/timex"
)

type trayMemberRow struct {
	ID         string  `json:"id"`
	TrayID     string  `json:"tray_id"`
	UserID     string  `json:"user_id"`
	JoinedAt   string  `json:"joined_at"`
	InvitedVia *string `json:"invited_via"`
}

func (r trayMemberRow) ToDomain() (domain.TrayMember, error) {
	ja, err := timex.ParseRFC3339OrNano(r.JoinedAt)
	if err != nil {
		return domain.TrayMember{}, fmt.Errorf("postgrest: tray_member joined_at: %w", err)
	}
	return domain.TrayMember{
		ID:         strings.TrimSpace(r.ID),
		TrayID:     strings.TrimSpace(r.TrayID),
		UserID:     strings.TrimSpace(r.UserID),
		JoinedAt:   ja,
		InvitedVia: r.InvitedVia,
	}, nil
}

func parseTrayMemberRows(raw []byte) ([]domain.TrayMember, error) {
	var rows []trayMemberRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("postgrest: parse tray_members: %w", err)
	}
	out := make([]domain.TrayMember, 0, len(rows))
	for _, r := range rows {
		m, err := r.ToDomain()
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

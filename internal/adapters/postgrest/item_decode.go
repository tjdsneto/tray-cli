package postgrest

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

// itemRowWithTray is used for ListOutbox (embed trays for owner check).
type itemRowWithTray struct {
	itemRow
	Trays *struct {
		OwnerID string `json:"owner_id"`
	} `json:"trays"`
}

type itemRow struct {
	ID                string  `json:"id"`
	TrayID            string  `json:"tray_id"`
	SourceUserID      string  `json:"source_user_id"`
	Title             string  `json:"title"`
	Status            string  `json:"status"`
	DueDate           *string `json:"due_date"`
	SnoozeUntil       *string `json:"snooze_until"`
	DeclineReason     *string `json:"decline_reason"`
	CompletionMessage *string `json:"completion_message"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

func itemFromRow(r itemRow) (domain.Item, error) {
	ca, err := parseTime(r.CreatedAt)
	if err != nil {
		return domain.Item{}, fmt.Errorf("postgrest: item created_at: %w", err)
	}
	ua, err := parseTime(r.UpdatedAt)
	if err != nil {
		return domain.Item{}, fmt.Errorf("postgrest: item updated_at: %w", err)
	}
	var snooze *time.Time
	if r.SnoozeUntil != nil && strings.TrimSpace(*r.SnoozeUntil) != "" {
		t, err := parseTime(strings.TrimSpace(*r.SnoozeUntil))
		if err != nil {
			return domain.Item{}, fmt.Errorf("postgrest: item snooze_until: %w", err)
		}
		snooze = &t
	}
	return domain.Item{
		ID:                r.ID,
		TrayID:            r.TrayID,
		SourceUserID:      r.SourceUserID,
		Title:             r.Title,
		Status:            r.Status,
		DueDate:           r.DueDate,
		SnoozeUntil:       snooze,
		DeclineReason:     r.DeclineReason,
		CompletionMessage: r.CompletionMessage,
		CreatedAt:         ca,
		UpdatedAt:         ua,
	}, nil
}

func parseItemRows(raw []byte) ([]domain.Item, error) {
	var rows []itemRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("postgrest: parse items: %w", err)
	}
	out := make([]domain.Item, 0, len(rows))
	for _, r := range rows {
		it, err := itemFromRow(r)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, nil
}

// outboxDomainItems keeps items the viewer filed on trays they do not own (pure filter + map).
func outboxDomainItems(rows []itemRowWithTray, viewerUserID string) ([]domain.Item, error) {
	me := strings.TrimSpace(viewerUserID)
	var out []domain.Item
	for _, row := range rows {
		if row.Trays == nil || strings.TrimSpace(row.Trays.OwnerID) == "" {
			continue
		}
		if row.Trays.OwnerID == me {
			continue
		}
		it, err := itemFromRow(row.itemRow)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, nil
}

func parseCreatedItem(raw []byte) (domain.Item, error) {
	var rows []itemRow
	if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 {
		return itemFromRow(rows[0])
	}
	var one itemRow
	if err := json.Unmarshal(raw, &one); err == nil && one.ID != "" {
		return itemFromRow(one)
	}
	return domain.Item{}, fmt.Errorf("postgrest: parse create item response: %s", strings.TrimSpace(string(raw)))
}

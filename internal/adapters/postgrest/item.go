package postgrest

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/timex"
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
	SortOrder         int     `json:"sort_order"`
	SourceUserID      string  `json:"source_user_id"`
	Title             string  `json:"title"`
	Status            string  `json:"status"`
	DueDate           *string `json:"due_date"`
	SnoozeUntil       *string `json:"snooze_until"`
	DeclineReason     *string `json:"decline_reason"`
	CompletionMessage *string `json:"completion_message"`
	AcceptedAt        *string `json:"accepted_at"`
	DeclinedAt        *string `json:"declined_at"`
	CompletedAt       *string `json:"completed_at"`
	ArchivedAt        *string `json:"archived_at"`
	SnoozedAt         *string `json:"snoozed_at"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

func (r itemRow) ToDomain() (domain.Item, error) {
	ca, err := timex.ParseRFC3339OrNano(r.CreatedAt)
	if err != nil {
		return domain.Item{}, fmt.Errorf("postgrest: item created_at: %w", err)
	}
	ua, err := timex.ParseRFC3339OrNano(r.UpdatedAt)
	if err != nil {
		return domain.Item{}, fmt.Errorf("postgrest: item updated_at: %w", err)
	}
	var snooze *time.Time
	if r.SnoozeUntil != nil && strings.TrimSpace(*r.SnoozeUntil) != "" {
		t, err := timex.ParseRFC3339OrNano(strings.TrimSpace(*r.SnoozeUntil))
		if err != nil {
			return domain.Item{}, fmt.Errorf("postgrest: item snooze_until: %w", err)
		}
		snooze = &t
	}
	acceptedAt, err := optionalRFC3339Time(r.AcceptedAt)
	if err != nil {
		return domain.Item{}, fmt.Errorf("postgrest: item accepted_at: %w", err)
	}
	declinedAt, err := optionalRFC3339Time(r.DeclinedAt)
	if err != nil {
		return domain.Item{}, fmt.Errorf("postgrest: item declined_at: %w", err)
	}
	completedAt, err := optionalRFC3339Time(r.CompletedAt)
	if err != nil {
		return domain.Item{}, fmt.Errorf("postgrest: item completed_at: %w", err)
	}
	archivedAt, err := optionalRFC3339Time(r.ArchivedAt)
	if err != nil {
		return domain.Item{}, fmt.Errorf("postgrest: item archived_at: %w", err)
	}
	snoozedAt, err := optionalRFC3339Time(r.SnoozedAt)
	if err != nil {
		return domain.Item{}, fmt.Errorf("postgrest: item snoozed_at: %w", err)
	}
	return domain.Item{
		ID:                r.ID,
		TrayID:            r.TrayID,
		SortOrder:         r.SortOrder,
		SourceUserID:      r.SourceUserID,
		Title:             r.Title,
		Status:            r.Status,
		DueDate:           r.DueDate,
		SnoozeUntil:       snooze,
		DeclineReason:     r.DeclineReason,
		CompletionMessage: r.CompletionMessage,
		AcceptedAt:        acceptedAt,
		DeclinedAt:        declinedAt,
		CompletedAt:       completedAt,
		ArchivedAt:        archivedAt,
		SnoozedAt:         snoozedAt,
		CreatedAt:         ca,
		UpdatedAt:         ua,
	}, nil
}

func optionalRFC3339Time(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil, nil
	}
	parsed, err := timex.ParseRFC3339OrNano(t)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseItemRows(raw []byte) ([]domain.Item, error) {
	var rows []itemRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("postgrest: parse items: %w", err)
	}
	out := make([]domain.Item, 0, len(rows))
	for _, r := range rows {
		it, err := r.ToDomain()
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
		it, err := row.itemRow.ToDomain()
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
		return rows[0].ToDomain()
	}
	var one itemRow
	if err := json.Unmarshal(raw, &one); err == nil && one.ID != "" {
		return one.ToDomain()
	}
	return domain.Item{}, fmt.Errorf("postgrest: parse create item response: %s", strings.TrimSpace(string(raw)))
}

// addItemRequest is the JSON body for POST /rest/v1/items.
type addItemRequest struct {
	TrayID       string  `json:"tray_id"`
	SourceUserID string  `json:"source_user_id"`
	Title        string  `json:"title"`
	Status       string  `json:"status"`
	DueDate      *string `json:"due_date,omitempty"`
}

func newAddItemRequest(userID, trayID, title string, dueDate *string, trayOwnerID string) (addItemRequest, error) {
	if strings.TrimSpace(userID) == "" {
		return addItemRequest{}, fmt.Errorf("postgrest: session missing UserID (set after login)")
	}
	tid := strings.TrimSpace(trayID)
	if tid == "" {
		return addItemRequest{}, fmt.Errorf("postgrest: empty tray id")
	}
	tit := strings.TrimSpace(title)
	if tit == "" {
		return addItemRequest{}, fmt.Errorf("postgrest: empty item title")
	}
	owner := strings.TrimSpace(trayOwnerID)
	if owner == "" {
		return addItemRequest{}, fmt.Errorf("postgrest: empty tray owner id")
	}
	st := "pending"
	if strings.EqualFold(strings.TrimSpace(userID), owner) {
		st = "accepted"
	}
	req := addItemRequest{
		TrayID:       tid,
		SourceUserID: strings.TrimSpace(userID),
		Title:        tit,
		Status:       st,
	}
	if dueDate != nil && strings.TrimSpace(*dueDate) != "" {
		d := strings.TrimSpace(*dueDate)
		req.DueDate = &d
	}
	return req, nil
}

func itemPatchBody(p domain.ItemPatch) (map[string]any, error) {
	body := map[string]any{}
	if p.Status != nil {
		body["status"] = strings.TrimSpace(*p.Status)
	}
	if p.SortOrder != nil {
		body["sort_order"] = *p.SortOrder
	}
	if p.DeclineReason != nil {
		body["decline_reason"] = strings.TrimSpace(*p.DeclineReason)
	}
	if p.CompletionMessage != nil {
		body["completion_message"] = strings.TrimSpace(*p.CompletionMessage)
	}
	if p.DueDate != nil {
		body["due_date"] = strings.TrimSpace(*p.DueDate)
	}
	if p.SnoozeUntil != nil {
		body["snooze_until"] = p.SnoozeUntil.UTC().Format(time.RFC3339Nano)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("postgrest: empty item patch")
	}
	return body, nil
}

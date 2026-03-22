package postgrest

import (
	"fmt"
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

// buildAddItemBody validates inputs and builds the JSON body for POST /items.
func buildAddItemBody(userID, trayID, title string, dueDate *string) (map[string]any, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("postgrest: session missing UserID (set after login)")
	}
	tid := strings.TrimSpace(trayID)
	if tid == "" {
		return nil, fmt.Errorf("postgrest: empty tray id")
	}
	tit := strings.TrimSpace(title)
	if tit == "" {
		return nil, fmt.Errorf("postgrest: empty item title")
	}
	body := map[string]any{
		"tray_id":        tid,
		"source_user_id": strings.TrimSpace(userID),
		"title":          tit,
		"status":         "pending",
	}
	if dueDate != nil && strings.TrimSpace(*dueDate) != "" {
		body["due_date"] = strings.TrimSpace(*dueDate)
	}
	return body, nil
}

func itemPatchBody(p domain.ItemPatch) (map[string]any, error) {
	body := map[string]any{}
	if p.Status != nil {
		body["status"] = strings.TrimSpace(*p.Status)
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

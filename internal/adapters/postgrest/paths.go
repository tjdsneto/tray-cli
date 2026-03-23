package postgrest

import (
	"net/url"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

const itemSelectColumns = "id,tray_id,source_user_id,title,status,due_date,snooze_until,decline_reason,completion_message,created_at,updated_at"

// itemsCreatePath is the POST /rest/v1/items path with Prefer: return=representation.
func itemsCreatePath() string {
	q := url.Values{}
	q.Set("select", itemSelectColumns)
	return "/rest/v1/items?" + q.Encode()
}

// itemsListPath is the GET /rest/v1/items path for List.
func itemsListPath(q domain.ListItemsQuery) string {
	u := url.Values{}
	u.Set("select", itemSelectColumns)
	if strings.TrimSpace(q.ItemID) != "" {
		u.Set("id", "eq."+strings.TrimSpace(q.ItemID))
	}
	if strings.TrimSpace(q.TrayID) != "" {
		u.Set("tray_id", "eq."+strings.TrimSpace(q.TrayID))
	}
	if strings.TrimSpace(q.Status) != "" {
		u.Set("status", "eq."+strings.TrimSpace(q.Status))
	}
	order := "created_at.desc"
	if strings.EqualFold(strings.TrimSpace(q.OrderCreated), "asc") {
		order = "created_at.asc"
	}
	u.Set("order", order)
	return "/rest/v1/items?" + u.Encode()
}

// itemsOutboxPath is the GET path for ListOutbox (items filed by viewer on others' trays).
func itemsOutboxPath(userID string) string {
	u := url.Values{}
	u.Set("select", itemSelectColumns+",trays(owner_id)")
	u.Set("source_user_id", "eq."+strings.TrimSpace(userID))
	u.Set("order", "created_at.desc")
	return "/rest/v1/items?" + u.Encode()
}

// itemsPatchPath is the PATCH /rest/v1/items path scoped by item id.
func itemsPatchPath(itemID string) string {
	id := strings.TrimSpace(itemID)
	q := url.Values{}
	q.Set("id", "eq."+id)
	return "/rest/v1/items?" + q.Encode()
}

// itemsDeletePath is the DELETE /rest/v1/items path scoped by item id.
func itemsDeletePath(itemID string) string { return itemsPatchPath(itemID) }

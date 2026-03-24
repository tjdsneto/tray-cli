package postgrest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/adapters/postgrest/pghttp"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

type itemService struct {
	pg *pghttp.Client
}

func newItemService(pg *pghttp.Client) *itemService {
	return &itemService{pg: pg}
}

var _ domain.ItemService = (*itemService)(nil)

func (s *itemService) Add(ctx context.Context, sess domain.Session, trayID, title string, dueDate *string) (*domain.Item, error) {
	body, err := newAddItemRequest(sess.UserID, trayID, title, dueDate)
	if err != nil {
		return nil, err
	}
	path := itemsCreatePath()
	raw, err := s.pg.Request(ctx, sess.AccessToken, http.MethodPost, path, body, pghttp.PreferRepresentation())
	if err != nil {
		return nil, err
	}
	it, err := parseCreatedItem(raw)
	if err != nil {
		return nil, err
	}
	return &it, nil
}

func (s *itemService) List(ctx context.Context, sess domain.Session, q domain.ListItemsQuery) ([]domain.Item, error) {
	path := itemsListPath(q)
	raw, err := s.pg.Request(ctx, sess.AccessToken, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return parseItemRows(raw)
}

func (s *itemService) ListOutbox(ctx context.Context, sess domain.Session) ([]domain.Item, error) {
	if strings.TrimSpace(sess.UserID) == "" {
		return nil, fmt.Errorf("postgrest: session missing UserID (set after login)")
	}
	path := itemsOutboxPath(sess.UserID)
	raw, err := s.pg.Request(ctx, sess.AccessToken, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	var rows []itemRowWithTray
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("postgrest: parse outbox items: %w", err)
	}
	return outboxDomainItems(rows, sess.UserID)
}

func (s *itemService) Update(ctx context.Context, sess domain.Session, itemID string, patch domain.ItemPatch) error {
	if strings.TrimSpace(sess.UserID) == "" {
		return fmt.Errorf("postgrest: session missing UserID (set after login)")
	}
	id := strings.TrimSpace(itemID)
	if id == "" {
		return fmt.Errorf("postgrest: empty item id")
	}
	body, err := itemPatchBody(patch)
	if err != nil {
		return err
	}
	path := itemsPatchPath(id)
	_, err = s.pg.Request(ctx, sess.AccessToken, http.MethodPatch, path, body, nil)
	return err
}

func (s *itemService) Delete(ctx context.Context, sess domain.Session, itemID string) error {
	if strings.TrimSpace(sess.UserID) == "" {
		return fmt.Errorf("postgrest: session missing UserID (set after login)")
	}
	id := strings.TrimSpace(itemID)
	if id == "" {
		return fmt.Errorf("postgrest: empty item id")
	}
	path := itemsDeletePath(id)
	_, err := s.pg.Request(ctx, sess.AccessToken, http.MethodDelete, path, nil, nil)
	return err
}

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
	if q.UpdatedAfter != nil && !q.UpdatedAfter.IsZero() {
		u.Set("updated_at", "gt."+q.UpdatedAfter.UTC().Format(time.RFC3339Nano))
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

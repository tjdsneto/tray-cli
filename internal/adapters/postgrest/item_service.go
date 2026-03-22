package postgrest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

type itemService struct {
	pg *client
}

func newItemService(pg *client) *itemService {
	return &itemService{pg: pg}
}

var _ domain.ItemService = (*itemService)(nil)

const itemSelect = "id,tray_id,source_user_id,title,status,due_date,snooze_until,decline_reason,completion_message,created_at,updated_at"

func (s *itemService) Add(ctx context.Context, sess domain.Session, trayID, title string, dueDate *string) (*domain.Item, error) {
	if strings.TrimSpace(sess.UserID) == "" {
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
		"source_user_id": sess.UserID,
		"title":          tit,
		"status":         "pending",
	}
	if dueDate != nil && strings.TrimSpace(*dueDate) != "" {
		body["due_date"] = strings.TrimSpace(*dueDate)
	}
	q := url.Values{}
	q.Set("select", itemSelect)
	path := "/rest/v1/items?" + q.Encode()
	raw, err := s.pg.request(ctx, sess, http.MethodPost, path, body, hdrPreferRepresentation())
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
	u := url.Values{}
	u.Set("select", itemSelect)
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
	path := "/rest/v1/items?" + u.Encode()
	raw, err := s.pg.request(ctx, sess, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return parseItemRows(raw)
}

func (s *itemService) ListOutbox(ctx context.Context, sess domain.Session) ([]domain.Item, error) {
	if strings.TrimSpace(sess.UserID) == "" {
		return nil, fmt.Errorf("postgrest: session missing UserID (set after login)")
	}
	u := url.Values{}
	u.Set("select", itemSelect+",trays(owner_id)")
	u.Set("source_user_id", "eq."+sess.UserID)
	u.Set("order", "created_at.desc")
	path := "/rest/v1/items?" + u.Encode()
	raw, err := s.pg.request(ctx, sess, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	var rows []itemRowWithTray
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("postgrest: parse outbox items: %w", err)
	}
	me := strings.TrimSpace(sess.UserID)
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
	q := url.Values{}
	q.Set("id", "eq."+id)
	path := "/rest/v1/items?" + q.Encode()
	_, err = s.pg.request(ctx, sess, http.MethodPatch, path, body, nil)
	return err
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

func (s *itemService) Delete(ctx context.Context, sess domain.Session, itemID string) error {
	return domain.ErrNotImplemented
}

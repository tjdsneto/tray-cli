package postgrest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

type itemService struct {
	pg *client
}

func newItemService(pg *client) *itemService {
	return &itemService{pg: pg}
}

var _ domain.ItemService = (*itemService)(nil)

func (s *itemService) Add(ctx context.Context, sess domain.Session, trayID, title string, dueDate *string) (*domain.Item, error) {
	body, err := buildAddItemBody(sess.UserID, trayID, title, dueDate)
	if err != nil {
		return nil, err
	}
	path := itemsCreatePath()
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
	path := itemsListPath(q)
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
	path := itemsOutboxPath(sess.UserID)
	raw, err := s.pg.request(ctx, sess, http.MethodGet, path, nil, nil)
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
	_, err = s.pg.request(ctx, sess, http.MethodPatch, path, body, nil)
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
	_, err := s.pg.request(ctx, sess, http.MethodDelete, path, nil, nil)
	return err
}

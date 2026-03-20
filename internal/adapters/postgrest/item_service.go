package postgrest

import (
	"context"

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
	return nil, domain.ErrNotImplemented
}

func (s *itemService) List(ctx context.Context, sess domain.Session, q domain.ListItemsQuery) ([]domain.Item, error) {
	return nil, domain.ErrNotImplemented
}

func (s *itemService) ListOutbox(ctx context.Context, sess domain.Session) ([]domain.Item, error) {
	return nil, domain.ErrNotImplemented
}

func (s *itemService) Update(ctx context.Context, sess domain.Session, itemID string, patch domain.ItemPatch) error {
	return domain.ErrNotImplemented
}

func (s *itemService) Delete(ctx context.Context, sess domain.Session, itemID string) error {
	return domain.ErrNotImplemented
}

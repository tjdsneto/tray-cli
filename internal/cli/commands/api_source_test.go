package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/api"
	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/localtray"
)

func TestAPISource_localAddListAndComplete(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := newAPISource(func() string { return dir }, func() (domain.Services, domain.Session, error) {
		return domain.Services{}, domain.Session{}, errors.New("signed out")
	})
	item, err := src.Add(context.Background(), api.AddRequest{Tray: "notes", Title: "Draft"})
	require.NoError(t, err)
	require.Equal(t, "local", item.Scope)
	require.Equal(t, "open", item.Status)

	items, err := src.Items(context.Background(), "notes")
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Nil(t, items[0].Body)

	err = src.Act(context.Background(), item.ID, "complete", api.ActBody{})
	require.NoError(t, err)
	items, err = src.Items(context.Background(), "notes")
	require.NoError(t, err)
	require.Empty(t, api.SelectItems(items, "open"))

	done, err := src.Item(context.Background(), item.ID)
	require.NoError(t, err)
	require.Equal(t, "local", done.Scope)
	require.Equal(t, item.ID, done.ID)
	require.Equal(t, "done", done.Status)

	err = src.Act(context.Background(), item.ID, "accept", api.ActBody{})
	require.ErrorContains(t, err, "local items only support complete")

	err = src.Act(context.Background(), item.ID, "complete", api.ActBody{})
	require.NoError(t, err)
}

func TestAPISource_addOnOwnedTrayDoesNotCreateLocal(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	items := &stubItemService{}
	src := newAPISource(func() string { return dir }, func() (domain.Services, domain.Session, error) {
		return domain.Services{
			Trays: &stubTrayService{owned: []domain.Tray{{ID: "remote-1", Name: "Inbox"}}},
			Items: items,
		}, domain.Session{}, nil
	})

	item, err := src.Add(context.Background(), api.AddRequest{Tray: "remote-1", Title: "Hello"})
	require.NoError(t, err)
	require.Equal(t, "remote", item.Scope)
	require.Equal(t, "remote-1", items.trayID)
	require.Equal(t, "Hello", items.title)
	require.Nil(t, items.due)
	require.Nil(t, items.agent)
	require.Nil(t, items.body)

	idx, err := localtray.LoadIndex(localtray.LocalDir(dir))
	require.NoError(t, err)
	require.Empty(t, idx.Trays)

	trays, err := src.Trays(context.Background())
	require.NoError(t, err)
	require.Equal(t, []api.Tray{{
		Scope: "remote",
		Ref:   "remote-1",
		Name:  "Inbox",
	}}, trays)
}

func TestAPISource_blankItemID(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := newAPISource(func() string { return dir }, func() (domain.Services, domain.Session, error) {
		return domain.Services{}, domain.Session{}, errors.New("signed out")
	})
	_, err := src.Item(context.Background(), " ")
	require.ErrorIs(t, err, api.ErrNotFound)
}

func TestAPISource_remoteListFailureKeepsLocalTrays(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	local := newAPISource(func() string { return dir }, func() (domain.Services, domain.Session, error) {
		return domain.Services{}, domain.Session{}, errors.New("signed out")
	})
	_, err := local.Add(context.Background(), api.AddRequest{Tray: "notes", Title: "Draft"})
	require.NoError(t, err)

	src := newAPISource(func() string { return dir }, func() (domain.Services, domain.Session, error) {
		return domain.Services{
			Trays: &stubTrayService{err: errors.New("remote down")},
			Items: &stubItemService{},
		}, domain.Session{}, nil
	})
	trays, err := src.Trays(context.Background())
	require.NoError(t, err)
	require.Len(t, trays, 1)
	require.Equal(t, "local", trays[0].Scope)
	require.Equal(t, "notes", trays[0].Ref)
	require.Equal(t, "notes", trays[0].Name)
}

type stubTrayService struct {
	owned []domain.Tray
	err   error
}

func (s *stubTrayService) ListOwned(context.Context, domain.Session) ([]domain.Tray, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.owned, nil
}

func (s *stubTrayService) Create(context.Context, domain.Session, string, *string) (*domain.Tray, error) {
	panic("unexpected TrayService.Create")
}
func (s *stubTrayService) ListMine(context.Context, domain.Session) ([]domain.Tray, error) {
	panic("unexpected TrayService.ListMine")
}
func (s *stubTrayService) ListJoined(context.Context, domain.Session) ([]domain.Tray, error) {
	panic("unexpected TrayService.ListJoined")
}
func (s *stubTrayService) UpdateName(context.Context, domain.Session, string, string) error {
	panic("unexpected TrayService.UpdateName")
}
func (s *stubTrayService) Delete(context.Context, domain.Session, string) error {
	panic("unexpected TrayService.Delete")
}
func (s *stubTrayService) SetInviteToken(context.Context, domain.Session, string, *string) error {
	panic("unexpected TrayService.SetInviteToken")
}
func (s *stubTrayService) Join(context.Context, domain.Session, string) (string, error) {
	panic("unexpected TrayService.Join")
}
func (s *stubTrayService) ListMembers(context.Context, domain.Session, string) ([]domain.TrayMember, error) {
	panic("unexpected TrayService.ListMembers")
}
func (s *stubTrayService) RemoveMember(context.Context, domain.Session, string, string) error {
	panic("unexpected TrayService.RemoveMember")
}
func (s *stubTrayService) Leave(context.Context, domain.Session, string) error {
	panic("unexpected TrayService.Leave")
}

type stubItemService struct {
	trayID string
	title  string
	due    *string
	agent  *string
	body   *string
}

func (s *stubItemService) Add(_ context.Context, _ domain.Session, trayID, title string, dueDate *string, agentSessionID *string, body *string) (*domain.Item, error) {
	s.trayID = trayID
	s.title = title
	s.due = dueDate
	s.agent = agentSessionID
	s.body = body
	return &domain.Item{
		ID:        "item-1",
		TrayID:    trayID,
		Title:     title,
		Status:    "pending",
		CreatedAt: time.Unix(1, 0).UTC(),
		Body:      body,
	}, nil
}

func (s *stubItemService) List(context.Context, domain.Session, domain.ListItemsQuery) ([]domain.Item, error) {
	return nil, nil
}

func (s *stubItemService) ListOutbox(context.Context, domain.Session) ([]domain.Item, error) {
	panic("unexpected ItemService.ListOutbox")
}
func (s *stubItemService) Update(context.Context, domain.Session, string, domain.ItemPatch) error {
	panic("unexpected ItemService.Update")
}
func (s *stubItemService) Delete(context.Context, domain.Session, string) error {
	panic("unexpected ItemService.Delete")
}
func (s *stubItemService) MoveUp(context.Context, domain.Session, string) error {
	panic("unexpected ItemService.MoveUp")
}
func (s *stubItemService) MoveDown(context.Context, domain.Session, string) error {
	panic("unexpected ItemService.MoveDown")
}

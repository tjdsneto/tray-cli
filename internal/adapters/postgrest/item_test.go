package postgrest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestItemRow_ToDomain(t *testing.T) {
	t.Parallel()
	it, err := (itemRow{
		ID: "i1", TrayID: "t1", SourceUserID: "u1",
		Title: "x", Status: "pending",
		CreatedAt: "2026-03-20T12:00:00Z",
		UpdatedAt: "2026-03-20T12:00:00Z",
	}).ToDomain()
	require.NoError(t, err)
	require.Equal(t, "i1", it.ID)
	require.Equal(t, "pending", it.Status)
}

func TestItemRow_ToDomain_statusTimestamps(t *testing.T) {
	t.Parallel()
	acc := "2026-03-22T15:00:00Z"
	it, err := (itemRow{
		ID: "i1", TrayID: "t1", SourceUserID: "u1",
		Title: "x", Status: "accepted",
		AcceptedAt: &acc,
		CreatedAt:  "2026-03-20T12:00:00Z",
		UpdatedAt:  "2026-03-22T15:00:00Z",
	}).ToDomain()
	require.NoError(t, err)
	require.NotNil(t, it.AcceptedAt)
	require.Equal(t, 2026, it.AcceptedAt.UTC().Year())
}

func TestItemRow_ToDomain_snoozeUntil(t *testing.T) {
	t.Parallel()
	s := "2026-03-21T10:00:00Z"
	it, err := (itemRow{
		ID: "i1", TrayID: "t1", SourceUserID: "u1",
		Title: "x", Status: "snoozed",
		SnoozeUntil: &s,
		CreatedAt:   "2026-03-20T12:00:00Z",
		UpdatedAt:   "2026-03-20T12:00:00Z",
	}).ToDomain()
	require.NoError(t, err)
	require.NotNil(t, it.SnoozeUntil)
}

func TestParseItemRows_invalidJSON(t *testing.T) {
	_, err := parseItemRows([]byte(`not-json`))
	require.Error(t, err)
}

func TestParseCreatedItem_object(t *testing.T) {
	raw := []byte(`{"id":"a","tray_id":"t","source_user_id":"u","title":"hi","status":"pending","created_at":"2026-03-20T12:00:00Z","updated_at":"2026-03-20T12:00:00Z"}`)
	it, err := parseCreatedItem(raw)
	require.NoError(t, err)
	require.Equal(t, "a", it.ID)
}

func TestParseCreatedItem_bad(t *testing.T) {
	_, err := parseCreatedItem([]byte(`{}`))
	require.Error(t, err)
}

func TestOutboxDomainItems_filtersOwnerAndNilTray(t *testing.T) {
	t.Parallel()
	rows := []itemRowWithTray{
		{itemRow: itemRow{ID: "1", TrayID: "t1", SourceUserID: "me", Title: "a", Status: "pending", CreatedAt: "2026-03-20T12:00:00Z", UpdatedAt: "2026-03-20T12:00:00Z"}, Trays: nil},
		{itemRow: itemRow{ID: "2", TrayID: "t2", SourceUserID: "me", Title: "b", Status: "pending", CreatedAt: "2026-03-20T12:00:00Z", UpdatedAt: "2026-03-20T12:00:00Z"}, Trays: &struct {
			OwnerID string `json:"owner_id"`
		}{OwnerID: "me"}},
		{itemRow: itemRow{ID: "3", TrayID: "t3", SourceUserID: "me", Title: "c", Status: "pending", CreatedAt: "2026-03-20T12:00:00Z", UpdatedAt: "2026-03-20T12:00:00Z"}, Trays: &struct {
			OwnerID string `json:"owner_id"`
		}{OwnerID: "owner"}},
	}
	out, err := outboxDomainItems(rows, "me")
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "3", out[0].ID)
}

func TestNewAddItemRequest_ok_pendingWhenNotOwner(t *testing.T) {
	t.Parallel()
	due := "2026-03-21"
	req, err := newAddItemRequest("u1", "t1", "hi", &due, "owner-2")
	require.NoError(t, err)
	require.Equal(t, "t1", req.TrayID)
	require.Equal(t, "u1", req.SourceUserID)
	require.Equal(t, "hi", req.Title)
	require.Equal(t, "pending", req.Status)
	require.NotNil(t, req.DueDate)
	require.Equal(t, due, *req.DueDate)
}

func TestNewAddItemRequest_acceptedWhenOwnerAdds(t *testing.T) {
	t.Parallel()
	req, err := newAddItemRequest("u1", "t1", "hi", nil, "u1")
	require.NoError(t, err)
	require.Equal(t, "accepted", req.Status)
}

func TestNewAddItemRequest_validation(t *testing.T) {
	t.Parallel()
	_, err := newAddItemRequest("", "t", "x", nil, "o")
	require.Error(t, err)
	_, err = newAddItemRequest("u", "", "x", nil, "o")
	require.Error(t, err)
	_, err = newAddItemRequest("u", "t", "", nil, "o")
	require.Error(t, err)
	_, err = newAddItemRequest("u", "t", "x", nil, "")
	require.Error(t, err)
}

func TestItemPatchBody_snooze(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	b, err := itemPatchBody(domain.ItemPatch{SnoozeUntil: &ts})
	require.NoError(t, err)
	require.Contains(t, b["snooze_until"].(string), "2026-03-21")
}

func TestItemPatchBody_empty(t *testing.T) {
	t.Parallel()
	_, err := itemPatchBody(domain.ItemPatch{})
	require.Error(t, err)
}

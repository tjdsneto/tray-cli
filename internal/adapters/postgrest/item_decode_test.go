package postgrest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestItemFromRow(t *testing.T) {
	t.Parallel()
	it, err := itemFromRow(itemRow{
		ID: "i1", TrayID: "t1", SourceUserID: "u1",
		Title: "x", Status: "pending",
		CreatedAt: "2026-03-20T12:00:00Z",
		UpdatedAt: "2026-03-20T12:00:00Z",
	})
	require.NoError(t, err)
	require.Equal(t, "i1", it.ID)
	require.Equal(t, "pending", it.Status)
}

func TestItemFromRow_snoozeUntil(t *testing.T) {
	t.Parallel()
	s := "2026-03-21T10:00:00Z"
	it, err := itemFromRow(itemRow{
		ID: "i1", TrayID: "t1", SourceUserID: "u1",
		Title: "x", Status: "snoozed",
		SnoozeUntil: &s,
		CreatedAt:   "2026-03-20T12:00:00Z",
		UpdatedAt:     "2026-03-20T12:00:00Z",
	})
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

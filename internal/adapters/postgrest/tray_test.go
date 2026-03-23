package postgrest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCreatedTray_arrayOrObject(t *testing.T) {
	row, err := parseCreatedTray([]byte(`[{"id":"x","owner_id":"o","name":"n","created_at":"2026-01-02T15:04:05Z"}]`))
	require.NoError(t, err)
	require.Equal(t, "x", row.ID)

	row, err = parseCreatedTray([]byte(`{"id":"y","owner_id":"o","name":"n2","created_at":"2026-01-02T15:04:05Z"}`))
	require.NoError(t, err)
	require.Equal(t, "y", row.ID)
}

func TestTrayRow_ToDomain_itemCount(t *testing.T) {
	tr, err := (trayRow{
		ID:        "id1",
		OwnerID:   "o",
		Name:      "t",
		CreatedAt: "2026-01-02T15:04:05Z",
		Items:     []trayItemsCount{{Count: 7}},
	}).ToDomain()
	require.NoError(t, err)
	require.Equal(t, 7, tr.ItemCount)
}

package api_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/api"
	"github.com/tjdsneto/tray-cli/internal/bar"
)

func ids(items []api.Item) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.ID
	}
	return out
}

func TestSnapshotFrom_includesClipboard(t *testing.T) {
	t.Parallel()
	barSnap := bar.BuildSnapshot([]bar.Item{
		{
			Scope:     bar.ScopeLocal,
			TrayRef:   "notes",
			TrayName:  "notes",
			ItemID:    "abc",
			Title:     "Draft",
			CreatedAt: time.Unix(1, 0).UTC(),
		},
	}, "")
	got := api.SnapshotFrom(barSnap)
	require.Equal(t, 1, got.Badge)
	require.Len(t, got.Sections, 1)
	require.Equal(t, "notes", got.Sections[0].Header)
	require.Len(t, got.Sections[0].Rows, 1)
	row := got.Sections[0].Rows[0]
	require.Equal(t, "abc", row.ItemID)
	require.Equal(t, "notes abc", row.Clipboard)
}

func TestFilterOpenAndSortNewest(t *testing.T) {
	t.Parallel()
	older := time.Unix(1, 0).UTC()
	newer := time.Unix(2, 0).UTC()
	items := []api.Item{
		{ID: "a", Scope: "local", Status: "done", CreatedAt: newer},
		{ID: "b", Scope: "local", Status: "open", CreatedAt: older},
		{ID: "c", Scope: "remote", Status: "pending", CreatedAt: newer},
		{ID: "d", Scope: "remote", Status: "completed", CreatedAt: newer},
		{ID: "e", Scope: "remote", Status: "snoozed", CreatedAt: older},
	}
	open := api.SelectItems(items, "open")
	require.Equal(t, []string{"c", "b", "e"}, ids(open))

	all := api.SelectItems(items, "all")
	require.Len(t, all, 5)
	require.Equal(t, "a", all[0].ID)
}

func TestValidateAdd(t *testing.T) {
	t.Parallel()
	require.Error(t, api.ValidateAdd("", "x"))
	require.Error(t, api.ValidateAdd("inbox", ""))
	require.Error(t, api.ValidateAdd("inbox", "   "))
	require.NoError(t, api.ValidateAdd("inbox", "Hello"))
}
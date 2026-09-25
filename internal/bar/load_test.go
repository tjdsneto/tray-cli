package bar_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/bar"
	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/localtray"
)

func TestItemsFromLocal(t *testing.T) {
	t.Parallel()
	rows := []localtray.ItemWithTray{{
		Item: localtray.Item{ID: "aabb", Title: "t", CreatedAt: time.Unix(1, 0).UTC()},
		Tray: localtray.TrayRecord{Kind: localtray.KindGlobal, Name: "inbox", ID: "global:inbox"},
	}}
	out := bar.ItemsFromLocal(rows)
	require.Len(t, out, 1)
	require.Equal(t, bar.ScopeLocal, out[0].Scope)
	require.Equal(t, "inbox", out[0].TrayRef)
	require.Equal(t, "inbox", out[0].TrayName)
	require.Equal(t, "aabb", out[0].ItemID)
	require.Equal(t, "t", out[0].Title)
}

func TestItemsFromLocal_emptyTrayName_usesDisplayLabel(t *testing.T) {
	t.Parallel()
	rows := []localtray.ItemWithTray{{
		Item: localtray.Item{ID: "x", Title: "y", CreatedAt: time.Unix(1, 0).UTC()},
		Tray: localtray.TrayRecord{Kind: localtray.KindGlobal, Name: "", ID: "global:inbox"},
	}}
	out := bar.ItemsFromLocal(rows)
	require.Len(t, out, 1)
	require.Equal(t, localtray.DisplayLabel("global:inbox"), out[0].TrayName)
}

func TestItemsFromRemote(t *testing.T) {
	t.Parallel()
	items := []domain.Item{{
		ID: "uuid-1", TrayID: "tray-uuid", Title: "p", CreatedAt: time.Unix(2, 0).UTC(),
	}}
	names := map[string]string{"tray-uuid": "work"}
	out := bar.ItemsFromRemote(items, names)
	require.Len(t, out, 1)
	require.Equal(t, "work", out[0].TrayRef)
	require.Equal(t, "work", out[0].TrayName)
	require.Equal(t, bar.ScopeRemote, out[0].Scope)
	require.Equal(t, "uuid-1", out[0].ItemID)
}

func TestItemsFromRemote_missingTrayName_fallsBackToTrayID(t *testing.T) {
	t.Parallel()
	items := []domain.Item{{
		ID: "uuid-1", TrayID: "tray-uuid", Title: "p", CreatedAt: time.Unix(2, 0).UTC(),
	}}
	out := bar.ItemsFromRemote(items, map[string]string{})
	require.Len(t, out, 1)
	require.Equal(t, "tray-uuid", out[0].TrayRef)
	require.Equal(t, "tray-uuid", out[0].TrayName)
}

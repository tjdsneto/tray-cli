package bar_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/bar"
)

func TestBuildSnapshot_groupsByTray_badgeAndNewestFirst(t *testing.T) {
	t.Parallel()
	older := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	in := []bar.Item{
		{Scope: bar.ScopeLocal, TrayRef: "inbox", TrayName: "inbox", ItemID: "loc1", Title: "old", CreatedAt: older},
		{Scope: bar.ScopeLocal, TrayRef: "inbox", TrayName: "inbox", ItemID: "loc2", Title: "new", CreatedAt: newer},
		{Scope: bar.ScopeRemote, TrayRef: "work", TrayName: "work", ItemID: "rem1", Title: "pending", CreatedAt: newer},
	}
	snap := bar.BuildSnapshot(in, "")
	require.Equal(t, 3, snap.Badge)
	require.Len(t, snap.Sections, 2)
	require.Equal(t, "inbox", snap.Sections[0].Header)
	require.Equal(t, "loc2", snap.Sections[0].Rows[0].ItemID)
	require.Equal(t, "work", snap.Sections[1].Header)
}

func TestBuildSnapshot_nameCollision_disambiguates(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	in := []bar.Item{
		{Scope: bar.ScopeLocal, TrayRef: "inbox", TrayName: "inbox", ItemID: "l1", Title: "a", CreatedAt: now},
		{Scope: bar.ScopeRemote, TrayRef: "inbox", TrayName: "inbox", ItemID: "r1", Title: "b", CreatedAt: now},
	}
	snap := bar.BuildSnapshot(in, "")
	require.Equal(t, 2, snap.Badge)
	require.Len(t, snap.Sections, 2)
	headers := []string{snap.Sections[0].Header, snap.Sections[1].Header}
	require.Contains(t, headers, "inbox · local")
	require.Contains(t, headers, "inbox · remote")
}

func TestClipboardLine(t *testing.T) {
	t.Parallel()
	require.Equal(t, "inbox abc123", bar.ClipboardLine(bar.Row{TrayRef: "inbox", ItemID: "abc123"}))
}

func TestClipboardLine_quotesWhitespaceTrayRef(t *testing.T) {
	t.Parallel()
	line := bar.ClipboardLine(bar.Row{TrayRef: "dir:/Users/me/My Projects", ItemID: "abc123"})
	require.Equal(t, `"dir:/Users/me/My Projects" abc123`, line)
}

func TestParseClipboardLine_unquoted(t *testing.T) {
	t.Parallel()
	ref, id, err := bar.ParseClipboardLine("inbox abc123")
	require.NoError(t, err)
	require.Equal(t, "inbox", ref)
	require.Equal(t, "abc123", id)
}

func TestParseClipboardLine_quoted(t *testing.T) {
	t.Parallel()
	ref, id, err := bar.ParseClipboardLine(`"dir:/Users/me/My Projects" abc123`)
	require.NoError(t, err)
	require.Equal(t, "dir:/Users/me/My Projects", ref)
	require.Equal(t, "abc123", id)
}

func TestParseClipboardLine_roundTrip(t *testing.T) {
	t.Parallel()
	row := bar.Row{TrayRef: "dir:/Users/me/My Projects", ItemID: "item99"}
	ref, id, err := bar.ParseClipboardLine(bar.ClipboardLine(row))
	require.NoError(t, err)
	require.Equal(t, row.TrayRef, ref)
	require.Equal(t, row.ItemID, id)
}

func TestRemoteItemsFromSnapshot(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	snap := bar.BuildSnapshot([]bar.Item{
		{Scope: bar.ScopeLocal, TrayRef: "inbox", TrayName: "inbox", ItemID: "l1", Title: "local", CreatedAt: now},
		{Scope: bar.ScopeRemote, TrayRef: "work", TrayName: "work", ItemID: "r1", Title: "remote", CreatedAt: now},
	}, "")
	remote := bar.RemoteItemsFromSnapshot(snap)
	require.Len(t, remote, 1)
	require.Equal(t, bar.ScopeRemote, remote[0].Scope)
	require.Equal(t, "work", remote[0].TrayRef)
	require.Equal(t, "r1", remote[0].ItemID)
}

func TestBuildSnapshot_remoteErrLine(t *testing.T) {
	t.Parallel()
	snap := bar.BuildSnapshot(nil, "remote unavailable")
	require.Equal(t, 0, snap.Badge)
	require.Equal(t, "remote unavailable", snap.RemoteStatus)
}

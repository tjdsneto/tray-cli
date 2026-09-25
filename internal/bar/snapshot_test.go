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

func TestBuildSnapshot_remoteErrLine(t *testing.T) {
	t.Parallel()
	snap := bar.BuildSnapshot(nil, "remote unavailable")
	require.Equal(t, 0, snap.Badge)
	require.Equal(t, "remote unavailable", snap.RemoteStatus)
}

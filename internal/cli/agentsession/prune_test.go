package agentsession_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/cli/agentsession"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestPrunable(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	idle := 7 * 24 * time.Hour
	old := now.Add(-8 * 24 * time.Hour)
	recent := now.Add(-2 * 24 * time.Hour)

	tray := domain.Tray{Name: "agent-session:abc", CreatedAt: old}
	require.True(t, agentsession.Prunable(tray, nil, now, idle), "empty + idle (CreatedAt only)")

	tray = domain.Tray{Name: "agent-session:abc", UpdatedAt: old}
	require.True(t, agentsession.Prunable(tray, nil, now, idle), "empty + idle (UpdatedAt)")

	tray = domain.Tray{Name: "agent-session:abc", CreatedAt: recent}
	require.False(t, agentsession.Prunable(tray, nil, now, idle), "empty but not idle (CreatedAt only)")

	tray = domain.Tray{Name: "agent-session:abc", UpdatedAt: recent}
	require.False(t, agentsession.Prunable(tray, nil, now, idle), "empty but not idle (UpdatedAt)")

	tray = domain.Tray{Name: "agent-session:abc", UpdatedAt: old}
	open := []domain.Item{{Status: "accepted", UpdatedAt: old}}
	require.False(t, agentsession.Prunable(tray, open, now, idle), "has open item")

	done := []domain.Item{{Status: "completed", UpdatedAt: old, CompletedAt: &old}}
	require.True(t, agentsession.Prunable(tray, done, now, idle), "only completed + idle")

	require.False(t, agentsession.Prunable(domain.Tray{Name: "inbox", UpdatedAt: old}, nil, now, idle))
}

func TestIsOpenStatus(t *testing.T) {
	t.Parallel()
	require.True(t, agentsession.IsOpenStatus("pending"))
	require.True(t, agentsession.IsOpenStatus("accepted"))
	require.True(t, agentsession.IsOpenStatus("snoozed"))
	require.False(t, agentsession.IsOpenStatus("completed"))
	require.False(t, agentsession.IsOpenStatus("declined"))
	require.False(t, agentsession.IsOpenStatus("archived"))
}

package triageui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestDetailLines_includesStatusTimes(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)
	it := domain.Item{
		ID: "i1", TrayID: "t1", SourceUserID: "u1",
		Title: "Hi", Status: "accepted",
		AcceptedAt: &ts,
		CreatedAt:  ts,
		UpdatedAt:  ts,
	}
	lines := detailLines(it, map[string]string{"t1": "inbox"}, nil, domain.Session{UserID: "u1"})
	var joined string
	for _, ln := range lines {
		joined += ln + "\n"
	}
	require.Contains(t, joined, "Accepted:")
	require.Contains(t, joined, "2026-03-22T10:00:00Z")
}

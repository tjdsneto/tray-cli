package output

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestSectionKeysInDisplayOrder(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	b := partitionItemsByStatus([]domain.Item{
		{ID: "1", Status: "completed", CreatedAt: ts},
		{ID: "2", Status: "accepted", CreatedAt: ts},
		{ID: "3", Status: "pending", CreatedAt: ts},
	})
	keys := sectionKeysInDisplayOrder(b)
	require.Equal(t, []string{"accepted", "pending", "completed"}, keys)
}

func TestSectionTitleForStatus(t *testing.T) {
	t.Parallel()
	require.Equal(t, "Accepted", sectionTitleForStatus("accepted"))
	require.Equal(t, "Zany", sectionTitleForStatus("zany"))
}

func TestGroupConsecutiveByTrayID(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	items := []domain.Item{
		{ID: "a", TrayID: "t1", SortOrder: 1, CreatedAt: ts},
		{ID: "b", TrayID: "t1", SortOrder: 2, CreatedAt: ts},
		{ID: "c", TrayID: "t2", SortOrder: 1, CreatedAt: ts},
	}
	g := groupConsecutiveByTrayID(items)
	require.Len(t, g, 2)
	require.Len(t, g[0], 2)
	require.Len(t, g[1], 1)
	require.Equal(t, "c", g[1][0].ID)
}

package output

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestTimeColumnHeader(t *testing.T) {
	t.Parallel()
	require.Equal(t, "COMPLETED ON", timeColumnHeader("completed"))
	require.Equal(t, "ADDED ON", timeColumnHeader("pending"))
}

func TestItemTimeDisplayForSection_completedUsesCompletedAt(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	created := now.Add(-48 * time.Hour)
	completed := now.Add(-2 * time.Hour)
	it := domain.Item{
		Status:      "completed",
		CreatedAt:   created,
		CompletedAt: &completed,
	}
	got := itemTimeDisplayForSection(it, "completed", now)
	require.Contains(t, got, "hour")
}

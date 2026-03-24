package output

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHumanizeTimeAgo(t *testing.T) {
	now := time.Date(2026, 3, 24, 18, 0, 0, 0, time.UTC)
	require.Equal(t, "just now", HumanizeTimeAgo(now.Add(-30*time.Second), now))
	require.Equal(t, "1 minute ago", HumanizeTimeAgo(now.Add(-1*time.Minute), now))
	require.Equal(t, "20 minutes ago", HumanizeTimeAgo(now.Add(-20*time.Minute), now))
	require.Equal(t, "1 hour ago", HumanizeTimeAgo(now.Add(-1*time.Hour), now))
	require.Equal(t, "3 hours ago", HumanizeTimeAgo(now.Add(-3*time.Hour), now))
	require.Equal(t, "yesterday", HumanizeTimeAgo(now.Add(-30*time.Hour), now))
}

func TestFormatSourceUser(t *testing.T) {
	require.Equal(t, "you", FormatSourceUser("user-1", "user-1"))
	require.Equal(t, "you", FormatSourceUser("user-1", "USER-1"))
	require.Equal(t, "a1b2c3d4", FormatSourceUser("a1b2c3d4-e5f6-7890-abcd-ef1234567890", ""))
	require.Equal(t, "—", FormatSourceUser("", "x"))
}

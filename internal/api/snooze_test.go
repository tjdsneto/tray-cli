package api_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/api"
)

func TestSnoozeUntil(t *testing.T) {
	t.Parallel()
	loc := time.FixedZone("UTC-3", -3*60*60)
	now := time.Date(2026, 6, 15, 18, 30, 0, 0, loc)

	hour, err := api.SnoozeUntil("hour", now)
	require.NoError(t, err)
	require.True(t, hour.Equal(now.Add(time.Hour)))

	tomorrow, err := api.SnoozeUntil("tomorrow", now)
	require.NoError(t, err)
	require.True(t, tomorrow.Equal(time.Date(2026, 6, 16, 9, 0, 0, 0, loc)))

	week, err := api.SnoozeUntil("week", now)
	require.NoError(t, err)
	require.True(t, week.Equal(time.Date(2026, 6, 22, 9, 0, 0, 0, loc)))

	_, err = api.SnoozeUntil("calendar", now)
	require.Error(t, err)
}
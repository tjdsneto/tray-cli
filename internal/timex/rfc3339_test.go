package timex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseRFC3339OrNano(t *testing.T) {
	t.Parallel()
	s := "2026-03-20T12:00:00Z"
	got, err := ParseRFC3339OrNano(s)
	require.NoError(t, err)
	require.Equal(t, time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC), got.UTC())

	got, err = ParseRFC3339OrNano("2026-03-20T12:00:00.123456789Z")
	require.NoError(t, err)
	require.Equal(t, 2026, got.UTC().Year())
}

func TestParseRFC3339OrNano_empty(t *testing.T) {
	t.Parallel()
	_, err := ParseRFC3339OrNano("")
	require.Error(t, err)
}

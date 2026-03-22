package postgrest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestBuildAddItemBody_ok(t *testing.T) {
	t.Parallel()
	due := "2026-03-21"
	b, err := buildAddItemBody("u1", "t1", "hi", &due)
	require.NoError(t, err)
	require.Equal(t, "t1", b["tray_id"])
	require.Equal(t, "u1", b["source_user_id"])
	require.Equal(t, "hi", b["title"])
	require.Equal(t, "pending", b["status"])
	require.Equal(t, due, b["due_date"])
}

func TestBuildAddItemBody_validation(t *testing.T) {
	t.Parallel()
	_, err := buildAddItemBody("", "t", "x", nil)
	require.Error(t, err)
	_, err = buildAddItemBody("u", "", "x", nil)
	require.Error(t, err)
	_, err = buildAddItemBody("u", "t", "", nil)
	require.Error(t, err)
}

func TestItemPatchBody_snooze(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	b, err := itemPatchBody(domain.ItemPatch{SnoozeUntil: &ts})
	require.NoError(t, err)
	require.Contains(t, b["snooze_until"].(string), "2026-03-21")
}

func TestItemPatchBody_empty(t *testing.T) {
	t.Parallel()
	_, err := itemPatchBody(domain.ItemPatch{})
	require.Error(t, err)
}

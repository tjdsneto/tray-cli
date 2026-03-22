package postgrest

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestItemsListPath_filtersAndOrder(t *testing.T) {
	t.Parallel()
	p := itemsListPath(domain.ListItemsQuery{
		ItemID:       "i1",
		TrayID:       "t1",
		Status:       "pending",
		OrderCreated: "asc",
	})
	require.Contains(t, p, "id=eq.i1")
	require.Contains(t, p, "tray_id=eq.t1")
	require.Contains(t, p, "status=eq.pending")
	require.Contains(t, p, "order=created_at.asc")
}

func TestItemsListPath_defaultOrderDesc(t *testing.T) {
	t.Parallel()
	p := itemsListPath(domain.ListItemsQuery{})
	require.Contains(t, p, "order=created_at.desc")
}

func TestItemsCreatePath_select(t *testing.T) {
	t.Parallel()
	p := itemsCreatePath()
	q, err := url.ParseQuery(strings.TrimPrefix(p, "/rest/v1/items?"))
	require.NoError(t, err)
	require.Equal(t, itemSelectColumns, q.Get("select"))
}

func TestItemsOutboxPath(t *testing.T) {
	t.Parallel()
	p := itemsOutboxPath("user-1")
	require.Contains(t, p, "source_user_id=eq.user-1")
	require.Contains(t, p, "trays%28owner_id%29") // trays(owner_id) encoded
}

func TestItemsPatchPath(t *testing.T) {
	t.Parallel()
	p := itemsPatchPath("  abc  ")
	require.Contains(t, p, "id=eq.abc")
}

package output

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestWriteItems_empty(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteItems(&buf, nil, nil, "", nil, FormatTable))
	require.Contains(t, buf.String(), "No items")
}

func TestWriteItems_markdown(t *testing.T) {
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	items := []domain.Item{{
		ID: "11111111-1111-1111-1111-111111111111", TrayID: "t1", Title: "Do", Status: "pending",
		CreatedAt: ts, UpdatedAt: ts, SourceUserID: "u",
	}}
	var buf bytes.Buffer
	require.NoError(t, WriteItems(&buf, items, map[string]string{"t1": "inbox"}, "u", nil, FormatMarkdown))
	out := buf.String()
	require.Contains(t, out, "### Pending")
	require.Contains(t, out, "#### inbox")
	require.Contains(t, out, "| ORD | id | Title |")
	require.NotContains(t, out, "| Tray |")
	require.Contains(t, out, "`11111111-1111-1111-1111-111111111111`")
	require.Contains(t, out, "inbox")
}

func TestWriteItems_table(t *testing.T) {
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	items := []domain.Item{{
		ID: "11111111-1111-1111-1111-111111111111", TrayID: "t1", Title: "Do", Status: "pending",
		CreatedAt: ts, UpdatedAt: ts, SourceUserID: "u",
	}}
	var buf bytes.Buffer
	require.NoError(t, WriteItems(&buf, items, map[string]string{"t1": "inbox"}, "u", nil, FormatTable))
	out := buf.String()
	require.Contains(t, out, "Pending")
	require.Contains(t, out, "inbox")
	require.Contains(t, out, "Do")
	require.Contains(t, out, "you") // FormatSourceUser for self
	require.Contains(t, out, "inbox\n") // tray group title (dim on TTY)
	require.Contains(t, out, "· 11111111-1111-1111-1111-111111111111")
	require.Contains(t, out, "you") // meta line still lists contributor
}

func TestWriteItems_table_longTitleFullyShown(t *testing.T) {
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	needle := "MIDDLE_OF_VERY_LONG_TITLE_FOR_WRAP_TEST"
	long := "Re-review requested on PR #689 (Epic: March 20 journey work) " + needle + " https://example.com/pull/689"
	items := []domain.Item{{
		ID: "22222222-2222-2222-2222-222222222222", TrayID: "t1", Title: long, Status: "pending", SortOrder: 3,
		CreatedAt: ts, UpdatedAt: ts, SourceUserID: "u2",
	}}
	var buf bytes.Buffer
	require.NoError(t, WriteItems(&buf, items, map[string]string{"t1": "work"}, "self", map[string]string{"u2": "Fernando Duro"}, FormatTable))
	out := buf.String()
	require.Contains(t, out, needle, "full title must not be truncated")
	require.Contains(t, out, "work\n")
	require.Contains(t, out, "Fernando Duro")
	require.Contains(t, out, "· 22222222-2222-2222-2222-222222222222")
}

func TestWrapPlainTitle_wordWrap(t *testing.T) {
	lines := wrapPlainTitle("one two three four five", 10)
	require.Equal(t, []string{"one two", "three", "four five"}, lines)
}

func TestWrapPlainTitle_longToken(t *testing.T) {
	lines := wrapPlainTitle("abcdefghijklmnop", 8)
	require.Equal(t, []string{"abcdefgh", "ijklmnop"}, lines)
}

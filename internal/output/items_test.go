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
		ID: "i1", TrayID: "t1", Title: "Do", Status: "pending",
		CreatedAt: ts, UpdatedAt: ts, SourceUserID: "u",
	}}
	var buf bytes.Buffer
	require.NoError(t, WriteItems(&buf, items, map[string]string{"t1": "inbox"}, "u", nil, FormatMarkdown))
	require.Contains(t, buf.String(), "### Pending")
	require.Contains(t, buf.String(), "|")
	require.Contains(t, buf.String(), "inbox")
}

func TestWriteItems_table(t *testing.T) {
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	items := []domain.Item{{
		ID: "i1", TrayID: "t1", Title: "Do", Status: "pending",
		CreatedAt: ts, UpdatedAt: ts, SourceUserID: "u",
	}}
	var buf bytes.Buffer
	require.NoError(t, WriteItems(&buf, items, map[string]string{"t1": "inbox"}, "u", nil, FormatTable))
	out := buf.String()
	require.Contains(t, out, "Pending")
	require.Contains(t, out, "inbox")
	require.Contains(t, out, "Do")
	require.Contains(t, out, "you") // FormatSourceUser for self
}

func TestWriteItems_table_longTitleFullyShown(t *testing.T) {
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	needle := "MIDDLE_OF_VERY_LONG_TITLE_FOR_WRAP_TEST"
	long := "Re-review requested on PR #689 (Epic: March 20 journey work) " + needle + " https://example.com/pull/689"
	items := []domain.Item{{
		ID: "i1", TrayID: "t1", Title: long, Status: "pending", SortOrder: 3,
		CreatedAt: ts, UpdatedAt: ts, SourceUserID: "u2",
	}}
	var buf bytes.Buffer
	require.NoError(t, WriteItems(&buf, items, map[string]string{"t1": "work"}, "self", map[string]string{"u2": "Fernando Duro"}, FormatTable))
	out := buf.String()
	require.Contains(t, out, needle, "full title must not be truncated")
	require.Contains(t, out, "   3  work · Fernando Duro ·")
}

func TestWrapPlainTitle_wordWrap(t *testing.T) {
	lines := wrapPlainTitle("one two three four five", 10)
	require.Equal(t, []string{"one two", "three", "four five"}, lines)
}

func TestWrapPlainTitle_longToken(t *testing.T) {
	lines := wrapPlainTitle("abcdefghijklmnop", 8)
	require.Equal(t, []string{"abcdefgh", "ijklmnop"}, lines)
}

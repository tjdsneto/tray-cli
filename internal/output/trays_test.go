package output

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestWriteTrays_table(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	trays := []domain.Tray{{Name: "a", ID: "111", ItemCount: 2, CreatedAt: ts}}
	var buf bytes.Buffer
	require.NoError(t, WriteTrays(&buf, trays, FormatTable, false, ""))
	require.Contains(t, buf.String(), "NAME")
	require.Contains(t, buf.String(), "ITEMS")
	require.Contains(t, buf.String(), "CREATED")
	require.Contains(t, buf.String(), "a")
	require.Contains(t, buf.String(), "2")
	require.NotContains(t, buf.String(), "111")
	require.Contains(t, buf.String(), "2026")
}

func TestWriteTrays_table_hints(t *testing.T) {
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	trays := []domain.Tray{{Name: "work", ItemCount: 0, CreatedAt: ts}}
	var buf bytes.Buffer
	require.NoError(t, WriteTrays(&buf, trays, FormatTable, true, ""))
	require.Contains(t, buf.String(), "Next steps")
	require.Contains(t, buf.String(), "tray add")
	require.Contains(t, buf.String(), "tray invite")
	require.Contains(t, buf.String(), "work")
}

func TestWriteTrays_table_hints_multipleTrays(t *testing.T) {
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	trays := []domain.Tray{
		{Name: "a", ItemCount: 1, CreatedAt: ts},
		{Name: "b", ItemCount: 0, CreatedAt: ts},
	}
	var buf bytes.Buffer
	require.NoError(t, WriteTrays(&buf, trays, FormatTable, true, ""))
	require.Contains(t, buf.String(), "<tray-name>")
	require.NotContains(t, buf.String(), "tray invite a")
}

func TestWriteTrays_json_includesID(t *testing.T) {
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	trays := []domain.Tray{{Name: "a", ID: "full-uuid-here", ItemCount: 0, CreatedAt: ts}}
	var buf bytes.Buffer
	require.NoError(t, WriteTrays(&buf, trays, FormatJSON, false, ""))
	require.Contains(t, buf.String(), "full-uuid-here")
	require.Contains(t, buf.String(), "item_count")
}

func TestWriteTrays_empty(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteTrays(&buf, nil, FormatTable, true, ""))
	s := buf.String()
	require.Contains(t, s, "No trays yet")
	require.Contains(t, s, "tray create")
	require.NotContains(t, s, "Next steps")
}

func TestWriteTrays_empty_noHints(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteTrays(&buf, nil, FormatTable, false, ""))
	require.Contains(t, buf.String(), "No trays.")
	require.NotContains(t, buf.String(), "tray create")
}

func TestWriteTrays_markdown_empty(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteTrays(&buf, nil, FormatMarkdown, false, ""))
	s := buf.String()
	require.Contains(t, s, "No trays yet")
	require.Contains(t, s, "tray create")
}

func TestWriteTrays_table_accessColumn(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	me := "user-her"
	trays := []domain.Tray{
		{Name: "inbox", ID: "1", OwnerID: me, ItemCount: 0, CreatedAt: ts},
		{Name: "work", ID: "2", OwnerID: "user-him", ItemCount: 1, CreatedAt: ts},
	}
	var buf bytes.Buffer
	require.NoError(t, WriteTrays(&buf, trays, FormatTable, false, me))
	s := buf.String()
	require.Contains(t, s, "ACCESS")
	require.Contains(t, s, "owner")
	require.Contains(t, s, "member")
	require.Contains(t, s, "work")
}

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
	require.NoError(t, WriteItems(&buf, nil, nil, FormatTable))
	require.Contains(t, buf.String(), "No items")
}

func TestWriteItems_table(t *testing.T) {
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	items := []domain.Item{{
		ID: "i1", TrayID: "t1", Title: "Do", Status: "pending",
		CreatedAt: ts, UpdatedAt: ts, SourceUserID: "u",
	}}
	var buf bytes.Buffer
	require.NoError(t, WriteItems(&buf, items, map[string]string{"t1": "inbox"}, FormatTable))
	require.Contains(t, buf.String(), "pending")
	require.Contains(t, buf.String(), "inbox")
}

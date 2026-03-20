package output

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestWriteTrays_table(t *testing.T) {
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	trays := []domain.Tray{{Name: "a", ID: "111", CreatedAt: ts}}
	var buf bytes.Buffer
	require.NoError(t, WriteTrays(&buf, trays, FormatTable))
	require.Contains(t, buf.String(), "NAME")
	require.Contains(t, buf.String(), "a")
	require.Contains(t, buf.String(), "111")
}

func TestWriteTrays_empty(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteTrays(&buf, nil, FormatTable))
	require.Contains(t, buf.String(), "No trays")
}

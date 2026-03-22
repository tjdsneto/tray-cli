package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteJoin_json(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteJoin(&buf, "tid", "inbox", FormatJSON))
	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	require.Equal(t, "tid", m["tray_id"])
	require.Equal(t, "inbox", m["name"])
}

func TestWriteJoin_human(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteJoin(&buf, "tid", "inbox", FormatTable))
	s := buf.String()
	require.Contains(t, s, `Joined tray "inbox"`)
	require.Contains(t, s, "tray ls")
}

func TestWriteJoin_markdown(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteJoin(&buf, "tid-1", "my tray", FormatMarkdown))
	s := buf.String()
	require.Contains(t, s, "|")
	require.Contains(t, s, `tid-1`)
	require.Contains(t, s, "my tray")
}

func TestWriteJoin_markdown_emptyName(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteJoin(&buf, "tid-1", "", FormatMarkdown))
	require.Contains(t, buf.String(), "—")
}

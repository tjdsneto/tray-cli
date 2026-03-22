package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteInvite_json(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteInvite(&buf, "t", "tok", FormatJSON))
	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	require.Equal(t, "tok", m["invite_token"])
}

func TestWriteInvite_human(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteInvite(&buf, "inbox", "secret", FormatTable))
	s := buf.String()
	require.Contains(t, s, `inbox`)
	require.Contains(t, s, "secret")
	require.Contains(t, s, "tray join")
}

func TestWriteInvite_markdown_escapesPipe(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteInvite(&buf, "a|b", "t|k", FormatMarkdown))
	s := buf.String()
	require.Contains(t, s, `\|`)
	require.Contains(t, s, "`tray join")
}

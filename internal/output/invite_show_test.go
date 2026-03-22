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

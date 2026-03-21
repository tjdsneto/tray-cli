package output

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShellQuoteTrayName(t *testing.T) {
	require.Equal(t, "work", shellQuoteTrayName("work"))
	require.Equal(t, `"my work"`, shellQuoteTrayName(`my work`))
	require.Equal(t, `""`, shellQuoteTrayName(""))
}

func TestWriteTrayHints_single(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteTrayHints(&buf, []string{"inbox"}))
	s := buf.String()
	require.Contains(t, s, `tray add "Task title" inbox`)
	require.Contains(t, s, "tray invite inbox")
}

package commands

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCmdTriage_Metadata(t *testing.T) {
	t.Parallel()
	c := cmdTriage()
	require.Equal(t, "triage [tray]", c.Use)
	require.Contains(t, c.Short, "triag")
	require.NotNil(t, c.RunE)
}

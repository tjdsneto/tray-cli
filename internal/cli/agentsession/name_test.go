package agentsession_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/cli/agentsession"
)

func TestTrayName_and_Parse(t *testing.T) {
	t.Parallel()
	name, err := agentsession.TrayName("  abc-123  ")
	require.NoError(t, err)
	require.Equal(t, "agent-session:abc-123", name)

	id, ok := agentsession.ParseTrayName(name)
	require.True(t, ok)
	require.Equal(t, "abc-123", id)

	_, err = agentsession.TrayName("")
	require.Error(t, err)
	_, err = agentsession.TrayName("   ")
	require.Error(t, err)

	_, ok = agentsession.ParseTrayName("inbox")
	require.False(t, ok)
	_, ok = agentsession.ParseTrayName("agent-session:")
	require.False(t, ok)
}

func TestIsAgentSessionTrayName(t *testing.T) {
	t.Parallel()
	require.True(t, agentsession.IsAgentSessionTrayName("agent-session:x"))
	require.False(t, agentsession.IsAgentSessionTrayName("Agent-Session:x")) // stored names are exact prefix
	require.False(t, agentsession.IsAgentSessionTrayName("work"))
}

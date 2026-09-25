package agentsession_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/cli/agentsession"
)

func TestFlagOrEnv(t *testing.T) {
	t.Setenv("TRAY_AGENT_SESSION_ID", "  from-env  ")

	require.Equal(t, "from-flag", agentsession.FlagOrEnv("  from-flag  ", "TRAY_AGENT_SESSION_ID"))
	require.Equal(t, "from-env", agentsession.FlagOrEnv("", "TRAY_AGENT_SESSION_ID"))
	require.Equal(t, "from-env", agentsession.FlagOrEnv("   ", "TRAY_AGENT_SESSION_ID"))

	t.Setenv("TRAY_AGENT_SESSION_ID", "")
	require.Equal(t, "", agentsession.FlagOrEnv("", "TRAY_AGENT_SESSION_ID"))
	require.Equal(t, "", agentsession.FlagOrEnv("   ", "TRAY_AGENT_SESSION_ID"))
}

func TestResolveListAgentSessionID(t *testing.T) {
	t.Parallel()

	id, active, err := agentsession.ResolveListAgentSessionID(false, "ignored", "env")
	require.NoError(t, err)
	require.False(t, active)
	require.Equal(t, "", id)

	id, active, err = agentsession.ResolveListAgentSessionID(true, "  sess-a  ", "env")
	require.NoError(t, err)
	require.True(t, active)
	require.Equal(t, "sess-a", id)

	id, active, err = agentsession.ResolveListAgentSessionID(true, "", "  from-env  ")
	require.NoError(t, err)
	require.True(t, active)
	require.Equal(t, "from-env", id)

	id, active, err = agentsession.ResolveListAgentSessionID(true, "   ", "")
	require.Error(t, err)
	require.True(t, active)
	require.Equal(t, "", id)
	require.Contains(t, err.Error(), "TRAY_AGENT_SESSION_ID")
}

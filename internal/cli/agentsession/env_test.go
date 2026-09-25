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

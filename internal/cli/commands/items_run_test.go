package commands

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/cli/agentsession"
)

func TestWithAddAliasDisplay_OverridesTrayNameWhenRefIsAlias(t *testing.T) {
	base := map[string]string{"tray-1": "work"}
	aliases := map[string]string{"tiago-work": "tray-1"}
	got := withAddAliasDisplay(base, "tiago-work", "tray-1", aliases)
	require.Equal(t, "tiago-work", got["tray-1"])
	require.Equal(t, "work", base["tray-1"])
}

func TestWithAddAliasDisplay_LeavesMapWhenRefIsNotAlias(t *testing.T) {
	base := map[string]string{"tray-1": "work"}
	aliases := map[string]string{"tiago-work": "tray-1"}
	got := withAddAliasDisplay(base, "work", "tray-1", aliases)
	require.Equal(t, "work", got["tray-1"])
}

func TestListAgentSessionID_BareFlagUsesSentinel(t *testing.T) {
	c := cmdList()
	c.RunE = func(*cobra.Command, []string) error { return nil }
	c.SetArgs([]string{"--agent-session-id"})
	require.NoError(t, c.Execute())

	val, err := c.Flags().GetString("agent-session-id")
	require.NoError(t, err)
	require.True(t, c.Flags().Changed("agent-session-id"))
	require.Equal(t, agentsession.ListAgentSessionIDFromEnv, val)

	id, active, err := agentsession.ResolveListAgentSessionID(true, val, "sess-from-env")
	require.NoError(t, err)
	require.True(t, active)
	require.Equal(t, "sess-from-env", id)
}

func TestListAgentSessionID_ExplicitIDSpaceSeparated(t *testing.T) {
	c := cmdList()
	var gotFlag, gotPos string
	c.RunE = func(cmd *cobra.Command, args []string) error {
		gotFlag, _ = cmd.Flags().GetString("agent-session-id")
		if len(args) == 1 {
			gotPos = args[0]
		}
		gotFlag, gotPos = agentsession.CoalesceListAgentSessionFlag(gotFlag, gotPos)
		return nil
	}
	c.SetArgs([]string{"--agent-session-id", "explicit-sess"})
	require.NoError(t, c.Execute())
	require.Equal(t, "explicit-sess", gotFlag)
	require.Equal(t, "", gotPos)

	id, active, err := agentsession.ResolveListAgentSessionID(true, gotFlag, "should-not-use")
	require.NoError(t, err)
	require.True(t, active)
	require.Equal(t, "explicit-sess", id)
}

func TestListAgentSessionID_ExplicitIDEqualsForm(t *testing.T) {
	c := cmdList()
	c.RunE = func(*cobra.Command, []string) error { return nil }
	c.SetArgs([]string{"--agent-session-id=explicit-sess"})
	require.NoError(t, c.Execute())

	val, err := c.Flags().GetString("agent-session-id")
	require.NoError(t, err)
	require.Equal(t, "explicit-sess", val)

	id, active, err := agentsession.ResolveListAgentSessionID(true, val, "should-not-use")
	require.NoError(t, err)
	require.True(t, active)
	require.Equal(t, "explicit-sess", id)
}

func TestAddAgentSessionID_BareFlagRequiresArgument(t *testing.T) {
	c := cmdAdd()
	c.SilenceErrors = true
	c.SilenceUsage = true
	c.RunE = func(*cobra.Command, []string) error { return nil }
	c.SetArgs([]string{"title", "--agent-session-id"})
	err := c.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "flag needs an argument")
}

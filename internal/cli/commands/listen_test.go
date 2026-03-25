package commands

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/cli/listenhook"
)

func TestCmdListen_Metadata(t *testing.T) {
	t.Parallel()
	c := cmdListen()
	require.Equal(t, "listen [tray]", c.Use)
	require.NotNil(t, c.RunE)
	f := c.Flags().Lookup("interval")
	require.NotNil(t, f)
	require.Equal(t, (10 * time.Second).String(), f.DefValue)
	require.NotNil(t, c.Flags().Lookup("hooks"))
	require.NotNil(t, c.Flags().Lookup("no-hooks"))
	require.NotNil(t, c.Flags().Lookup("quiet"))
	require.NotNil(t, c.Flags().Lookup("exec"))
	require.NotNil(t, c.Flags().Lookup("exec-pattern"))
	require.NotNil(t, c.Flags().Lookup("mode"))
	require.NotNil(t, c.Flags().Lookup("daemon"))
}

func TestParseExecEvents(t *testing.T) {
	t.Parallel()
	got, err := parseExecEvents("pending,completed accepted")
	require.NoError(t, err)
	require.Equal(t, []string{
		listenhook.EventItemPending,
		listenhook.EventItemCompleted,
		listenhook.EventItemAccepted,
	}, got)

	got, err = parseExecEvents("all")
	require.NoError(t, err)
	require.Equal(t, []string{
		listenhook.EventItemPending,
		listenhook.EventItemCompleted,
		listenhook.EventItemAccepted,
		listenhook.EventItemDeclined,
	}, got)

	_, err = parseExecEvents("bogus")
	require.Error(t, err)
}

func TestExecRules(t *testing.T) {
	t.Parallel()
	rules, err := execRules("echo hi", "declined")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.Equal(t, listenhook.EventItemDeclined, rules[0].Event)
	if runtime.GOOS == "windows" {
		require.Equal(t, []string{"cmd", "/C", "echo hi"}, rules[0].Command)
	} else {
		require.Equal(t, []string{"/bin/sh", "-c", "echo hi"}, rules[0].Command)
	}
}

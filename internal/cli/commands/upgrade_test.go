package commands

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRunUpgrade_UsesVersionAndPrintsSuccess(t *testing.T) {
	prev := runUpgradeScript
	t.Cleanup(func() { runUpgradeScript = prev })

	called := false
	runUpgradeScript = func(cmd *cobra.Command, version string) error {
		called = true
		require.Equal(t, "v0.1.1", version)
		return nil
	}

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	require.NoError(t, runUpgrade(cmd, "v0.1.1"))
	require.True(t, called)
	s := out.String()
	require.Contains(t, s, "Upgrading tray via install script (TRAY_VERSION=v0.1.1)")
	require.Contains(t, s, "Upgrade completed")
}

func TestRunUpgrade_EmptyVersionDefaultsToLatest(t *testing.T) {
	prev := runUpgradeScript
	t.Cleanup(func() { runUpgradeScript = prev })

	runUpgradeScript = func(cmd *cobra.Command, version string) error {
		require.Equal(t, "latest", version)
		return nil
	}

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	require.NoError(t, runUpgrade(cmd, ""))
}

func TestRunUpgrade_RejectsInvalidVersion(t *testing.T) {
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := runUpgrade(cmd, "v0.1.1 bad")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid --version value")
}

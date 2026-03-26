package commands

import (
	"testing"

	"github.com/stretchr/testify/require"
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

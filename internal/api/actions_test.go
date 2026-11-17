package api_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/api"
)

func TestActionsFor(t *testing.T) {
	t.Parallel()
	require.Equal(t, []string{"complete"}, api.ActionsFor("local", "open"))
	require.Equal(t, []string{"accept", "decline", "snooze"}, api.ActionsFor("remote", "pending"))
	require.Equal(t, []string{"complete", "archive"}, api.ActionsFor("remote", "accepted"))
	require.Equal(t, []string{"archive"}, api.ActionsFor("remote", "snoozed"))
	require.Empty(t, api.ActionsFor("local", "done"))
	require.Empty(t, api.ActionsFor("remote", "declined"))
	require.Empty(t, api.ActionsFor("local", "pending"))
}
package commands

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestPickSingleTrayID(t *testing.T) {
	t.Parallel()
	trays := []domain.Tray{
		{ID: "a", Name: "agent-session:x"},
		{ID: "b", Name: "other"},
	}
	id, err := pickSingleTrayID(trays, "agent-session:x")
	require.NoError(t, err)
	require.Equal(t, "a", id)

	id, err = pickSingleTrayID(trays, "missing")
	require.NoError(t, err)
	require.Equal(t, "", id)

	dupes := []domain.Tray{
		{ID: "1", Name: "agent-session:x"},
		{ID: "2", Name: "Agent-Session:x"},
	}
	_, err = pickSingleTrayID(dupes, "agent-session:x")
	require.Error(t, err)
}

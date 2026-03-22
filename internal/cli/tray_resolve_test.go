package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestFindTraysByNameFold(t *testing.T) {
	trays := []domain.Tray{
		{Name: "Inbox", ID: "1"},
		{Name: "work", ID: "2"},
	}
	require.Len(t, findTraysByNameFold(trays, "inbox"), 1)
	require.Len(t, findTraysByNameFold(trays, "WORK"), 1)
	require.Empty(t, findTraysByNameFold(trays, "nope"))
}

func TestPickTrayOrError(t *testing.T) {
	_, err := pickTrayOrError(nil, "x")
	require.Error(t, err)
	t1 := domain.Tray{ID: "a"}
	_, err = pickTrayOrError([]domain.Tray{t1, t1}, "x")
	require.Error(t, err)
	got, err := pickTrayOrError([]domain.Tray{t1}, "x")
	require.NoError(t, err)
	require.Equal(t, "a", got.ID)
}

package trayref

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
	require.Len(t, FindTraysByNameFold(trays, "inbox"), 1)
	require.Len(t, FindTraysByNameFold(trays, "WORK"), 1)
	require.Empty(t, FindTraysByNameFold(trays, "nope"))
}

func TestTrayIDFromRef(t *testing.T) {
	trays := []domain.Tray{{ID: "uuid-1", Name: "Inbox"}}
	id, err := TrayIDFromRef("uuid-1", nil, trays)
	require.NoError(t, err)
	require.Equal(t, "uuid-1", id)

	id, err = TrayIDFromRef("inbox", nil, trays)
	require.NoError(t, err)
	require.Equal(t, "uuid-1", id)

	id, err = TrayIDFromRef("boss", map[string]string{"boss": "remote-id"}, nil)
	require.NoError(t, err)
	require.Equal(t, "remote-id", id)

	_, err = TrayIDFromRef("  ", nil, trays)
	require.Error(t, err)
}

func TestPickTrayOrError(t *testing.T) {
	_, err := PickTrayOrError(nil, "x")
	require.Error(t, err)
	t1 := domain.Tray{ID: "a"}
	_, err = PickTrayOrError([]domain.Tray{t1, t1}, "x")
	require.Error(t, err)
	got, err := PickTrayOrError([]domain.Tray{t1}, "x")
	require.NoError(t, err)
	require.Equal(t, "a", got.ID)
}

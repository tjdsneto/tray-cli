package commands

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/api"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestAPISource_localAddListAndComplete(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := newAPISource(func() string { return dir }, func() (domain.Services, domain.Session, error) {
		return domain.Services{}, domain.Session{}, errors.New("signed out")
	})
	item, err := src.Add(context.Background(), api.AddRequest{Tray: "notes", Title: "Draft"})
	require.NoError(t, err)
	require.Equal(t, "local", item.Scope)
	require.Equal(t, "open", item.Status)

	items, err := src.Items(context.Background(), "notes")
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Nil(t, items[0].Body)

	err = src.Act(context.Background(), item.ID, "complete", api.ActBody{})
	require.NoError(t, err)
	items, err = src.Items(context.Background(), "notes")
	require.NoError(t, err)
	require.Empty(t, api.SelectItems(items, "open"))

	err = src.Act(context.Background(), item.ID, "accept", api.ActBody{})
	require.Error(t, err)
}

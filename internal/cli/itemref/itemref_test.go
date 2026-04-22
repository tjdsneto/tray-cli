package itemref

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestParseItemRef_fullWithHyphens(t *testing.T) {
	full, prefix, err := ParseItemRef("F07A9C2E-4D1B-4C3E-9F2A-1B2C3D4E5F66")
	require.NoError(t, err)
	require.Equal(t, "f07a9c2e-4d1b-4c3e-9f2a-1b2c3d4e5f66", full)
	require.Empty(t, prefix)
}

func TestParseItemRef_full32Hex(t *testing.T) {
	full, prefix, err := ParseItemRef("f07a9c2e4d1b4c3e9f2a1b2c3d4e5f66")
	require.NoError(t, err)
	require.Equal(t, "f07a9c2e-4d1b-4c3e-9f2a-1b2c3d4e5f66", full)
	require.Empty(t, prefix)
}

func TestParseItemRef_prefix(t *testing.T) {
	full, prefix, err := ParseItemRef("f07a9c2e")
	require.NoError(t, err)
	require.Empty(t, full)
	require.Equal(t, "f07a9c2e", prefix)
}

func TestParseItemRef_prefixTooShort(t *testing.T) {
	_, _, err := ParseItemRef("f07a9c")
	require.Error(t, err)
}

func TestParseItemRef_nonHex(t *testing.T) {
	_, _, err := ParseItemRef("f07a9c2g")
	require.Error(t, err)
}

func TestResolveAmongItems_unique(t *testing.T) {
	items := []domain.Item{
		{ID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"},
		{ID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"},
	}
	id, err := ResolveAmongItems(items, "aaaa")
	require.NoError(t, err)
	require.Equal(t, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", id)
}

func TestResolveAmongItems_ambiguous(t *testing.T) {
	items := []domain.Item{
		{ID: "f07a9c2e-1111-1111-1111-111111111111"},
		{ID: "f07a9c2e-2222-2222-2222-222222222222"},
	}
	_, err := ResolveAmongItems(items, "f07a9c2e")
	require.Error(t, err)
	require.Contains(t, err.Error(), "ambiguous")
}

func TestResolveAmongItems_none(t *testing.T) {
	_, err := ResolveAmongItems([]domain.Item{{ID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}}, "bbbbbbbb")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no item matches")
}

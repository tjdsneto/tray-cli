package bar_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/bar"
	"github.com/tjdsneto/tray-cli/internal/localtray"
)

func TestLocalTrayRef(t *testing.T) {
	t.Parallel()
	require.Equal(t, "inbox", bar.LocalTrayRef(localtray.TrayRecord{
		Kind: localtray.KindGlobal, Name: "inbox", ID: "global:inbox",
	}))
	require.Equal(t, "dir:/Users/me/proj", bar.LocalTrayRef(localtray.TrayRecord{
		Kind: localtray.KindDir, Path: "/Users/me/proj", ID: "dir:/Users/me/proj",
	}))
	require.Equal(t, "branch:abc:feat", bar.LocalTrayRef(localtray.TrayRecord{
		Kind: localtray.KindBranch, ID: "branch:abc:feat",
	}))
}

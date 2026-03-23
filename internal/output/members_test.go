package output

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestWriteTrayMembers_table(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	members := []domain.TrayMember{{
		ID: "1", TrayID: "t1", UserID: "u1", JoinedAt: ts,
	}}
	var buf bytes.Buffer
	require.NoError(t, WriteTrayMembers(&buf, "inbox", members, FormatTable))
	require.Contains(t, buf.String(), "u1")
	require.Contains(t, buf.String(), "USER_ID")
}

func TestWriteTrayMembers_json(t *testing.T) {
	ts := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	members := []domain.TrayMember{{ID: "1", TrayID: "t1", UserID: "u1", JoinedAt: ts}}
	var buf bytes.Buffer
	require.NoError(t, WriteTrayMembers(&buf, "inbox", members, FormatJSON))
	require.Contains(t, buf.String(), `"user_id"`)
	require.Contains(t, buf.String(), "u1")
}

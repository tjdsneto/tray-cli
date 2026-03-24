package postgrest

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProfilesListPath(t *testing.T) {
	p := profilesListPath([]string{"a", "b"})
	require.True(t, strings.HasPrefix(p, "/rest/v1/profiles?"))
	q, err := url.ParseQuery(strings.TrimPrefix(p, "/rest/v1/profiles?"))
	require.NoError(t, err)
	require.Equal(t, "id,email,full_name", q.Get("select"))
	require.Equal(t, `in.("a","b")`, q.Get("id"))

	u := "76d12c8f-e5f6-7890-abcd-ef1234567890"
	p2 := profilesListPath([]string{u})
	q2, err := url.ParseQuery(strings.TrimPrefix(p2, "/rest/v1/profiles?"))
	require.NoError(t, err)
	require.Equal(t, `in.("76d12c8f-e5f6-7890-abcd-ef1234567890")`, q2.Get("id"))
}

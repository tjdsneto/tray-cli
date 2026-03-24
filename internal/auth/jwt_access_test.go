package auth

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAccessTokenExpiry(t *testing.T) {
	// eyJhbGciOiJub25lIn0 = {"alg":"none"} ; eyJleHAiOjE3MDAwMDAwMDB9 = {"exp":1700000000}
	tok := "eyJhbGciOiJub25lIn0.eyJleHAiOjE3MDAwMDAwMDB9.sig"
	exp, ok := accessTokenExpiry(tok)
	require.True(t, ok)
	require.Equal(t, int64(1700000000), exp.Unix())
}

func TestAccessTokenNeedsRefresh(t *testing.T) {
	far := time.Now().Add(1 * time.Hour).Unix()
	tok := "eyJhbGciOiJub25lIn0." + mustB64JSON(t, map[string]any{"exp": float64(far)}) + ".x"
	require.False(t, accessTokenNeedsRefresh(tok, 2*time.Minute))

	near := time.Now().Add(30 * time.Second).Unix()
	tok2 := "eyJhbGciOiJub25lIn0." + mustB64JSON(t, map[string]any{"exp": float64(near)}) + ".x"
	require.True(t, accessTokenNeedsRefresh(tok2, 2*time.Minute))
}

func mustB64JSON(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	require.NoError(t, err)
	return base64.RawURLEncoding.EncodeToString(raw)
}

package commands

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandshakeLine(t *testing.T) {
	t.Parallel()
	line := handshakeLine(4321, "tok")
	var got struct {
		Port  int    `json:"port"`
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal([]byte(line), &got))
	require.Equal(t, 4321, got.Port)
	require.Equal(t, "tok", got.Token)
}

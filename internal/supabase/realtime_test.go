package supabase

import (
	"encoding/json"
	"errors"
	"net"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestNewRealtimeClient_websocketURL(t *testing.T) {
	t.Parallel()
	c, err := NewRealtimeClient("https://abc.supabase.co", "anon-key")
	require.NoError(t, err)
	require.Contains(t, c.wsURL, "wss://abc.supabase.co/realtime/v1/websocket")
	require.Contains(t, c.wsURL, "apikey=anon-key")
	require.Contains(t, c.wsURL, "vsn=1.0.0")
}

func TestIsRetryableRealtimeErr(t *testing.T) {
	t.Parallel()
	require.True(t, IsRetryableRealtimeErr(errors.New("read tcp: i/o timeout")))
	require.True(t, IsRetryableRealtimeErr(&net.OpError{Err: contextDeadlineError{}}))

	var timeout net.Error = timeoutError{}
	require.True(t, IsRetryableRealtimeErr(timeout))
	require.False(t, IsRetryableRealtimeErr(errors.New("bad handshake")))
	require.False(t, IsRetryableRealtimeErr(errors.New("no events requested")))
	require.False(t, IsRetryableRealtimeErr(nil))

	require.True(t, IsRetryableRealtimeErr(&websocket.CloseError{Code: websocket.CloseAbnormalClosure}))
	require.False(t, IsRetryableRealtimeErr(&websocket.CloseError{Code: websocket.CloseNormalClosure}))

	require.True(t, IsRetryableRealtimeErr(errors.New("realtime stream closed")))
}

func TestParseEnvelope_EventTypeShape(t *testing.T) {
	t.Parallel()
	payload := map[string]any{
		"data": map[string]any{
			"eventType": "INSERT",
			"new":       map[string]any{"id": "1"},
			"old":       map[string]any{},
		},
	}
	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	ch, ok := parseEnvelope(envelope{Event: "postgres_changes", Payload: raw})
	require.True(t, ok)
	require.Equal(t, "INSERT", ch.Type)
	require.Equal(t, "1", ch.New["id"])
}

func TestParseEnvelope_RecordShape(t *testing.T) {
	t.Parallel()
	payload := map[string]any{
		"data": map[string]any{
			"type":       "UPDATE",
			"record":     map[string]any{"id": "2"},
			"old_record": map[string]any{"id": "2"},
		},
	}
	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	ch, ok := parseEnvelope(envelope{Event: "postgres_changes", Payload: raw})
	require.True(t, ok)
	require.Equal(t, "UPDATE", ch.Type)
	require.Equal(t, "2", ch.New["id"])
	require.Equal(t, "2", ch.Old["id"])
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

type contextDeadlineError struct{}

func (contextDeadlineError) Error() string   { return "context deadline exceeded" }
func (contextDeadlineError) Timeout() bool { return true }

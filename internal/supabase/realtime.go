package supabase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Note: WebSocket auth matches hosted Realtime — anon apikey + vsn in the URL only.
// User JWT belongs in phx_join payload (joinPayload). Do not set Authorization/apikey
// on the HTTP upgrade; Supabase/Cloudflare may reject with 4xx → "bad handshake".

// Change is a realtime database change for public.items.
type Change struct {
	Type string         // INSERT | UPDATE | DELETE
	New  map[string]any // row after change (if present)
	Old  map[string]any // row before change (if present)
}

// RealtimeClient subscribes to Supabase Realtime over websockets.
type RealtimeClient struct {
	wsURL string
}

func NewRealtimeClient(rawBaseURL, anonKey string) (*RealtimeClient, error) {
	base := strings.TrimSpace(rawBaseURL)
	anonKey = strings.TrimSpace(anonKey)
	if base == "" {
		return nil, fmt.Errorf("supabase realtime: empty base URL")
	}
	if anonKey == "" {
		return nil, fmt.Errorf("supabase realtime: empty anon key")
	}
	u, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil {
		return nil, fmt.Errorf("supabase realtime: parse base URL: %w", err)
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	default:
		return nil, fmt.Errorf("supabase realtime: unsupported URL scheme %q", u.Scheme)
	}
	u.Path = "/realtime/v1/websocket"
	q := u.Query()
	q.Set("apikey", anonKey)
	q.Set("vsn", "1.0.0")
	u.RawQuery = q.Encode()
	return &RealtimeClient{wsURL: u.String()}, nil
}

func realtimeDialer() *websocket.Dialer {
	// DefaultDialer uses a 45s handshake; slow or lossy paths (Wi‑Fi, VPN) often need more.
	return &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 90 * time.Second,
		// Compression can confuse some intermediaries; Supabase text frames work without it.
		EnableCompression: false,
	}
}

// IsRetryableRealtimeErr reports whether a failed realtime connection may succeed on retry.
func IsRetryableRealtimeErr(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	switch {
	case strings.Contains(s, "no events requested"),
		strings.Contains(s, "missing access token"),
		strings.Contains(s, "empty base URL"),
		strings.Contains(s, "empty anon key"),
		strings.Contains(s, "unsupported URL scheme"):
		return false
	case strings.Contains(s, "bad handshake"):
		return false
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}
	if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
		return true
	}
	if strings.Contains(s, "realtime disconnected") || strings.Contains(s, "realtime stream closed") {
		return true
	}
	return strings.Contains(s, "timeout") ||
		strings.Contains(s, "i/o timeout") ||
		strings.Contains(s, "connection reset") ||
		strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "EOF") ||
		strings.Contains(s, "unexpected EOF") ||
		strings.Contains(s, "close 1006") ||
		strings.Contains(s, "abnormal closure")
}

func dialRealtime(ctx context.Context, d *websocket.Dialer, urlStr string, hdr http.Header) (*websocket.Conn, *http.Response, error) {
	backoffs := []time.Duration{0, 400 * time.Millisecond, 900 * time.Millisecond}
	var lastErr error
	for i, wait := range backoffs {
		if i > 0 {
			select {
			case <-ctx.Done():
				return nil, nil, ctx.Err()
			case <-time.After(wait):
			}
		}
		conn, resp, err := d.DialContext(ctx, urlStr, hdr)
		if err == nil {
			return conn, resp, nil
		}
		lastErr = err
		if !IsRetryableRealtimeErr(err) {
			break
		}
	}
	return nil, nil, lastErr
}

// SubscribeItems subscribes to public.items changes for the current JWT.
func (c *RealtimeClient) SubscribeItems(ctx context.Context, accessToken string) (<-chan Change, <-chan error, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, nil, fmt.Errorf("supabase realtime: missing access token")
	}
	conn, _, err := dialRealtime(ctx, realtimeDialer(), c.wsURL, nil)
	if err != nil {
		return nil, nil, err
	}
	out := make(chan Change, 32)
	errs := make(chan error, 1)

	joinRef := "1"
	join := map[string]any{
		"topic":    "realtime:public:items",
		"event":    "phx_join",
		"payload":  joinPayload(accessToken),
		"ref":      joinRef,
		"join_ref": joinRef,
	}
	if err := conn.WriteJSON(join); err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	go func() {
		defer close(out)
		defer close(errs)
		defer conn.Close()
		_ = conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		})
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()

		readDone := make(chan struct{})
		go func() {
			defer close(readDone)
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					select {
					case errs <- err:
					default:
					}
					return
				}
				// Extend on every frame: Phoenix acks heartbeats as JSON, not WS pongs, so
				// relying only on PongHandler could let the read deadline fire during an otherwise healthy session.
				_ = conn.SetReadDeadline(time.Now().Add(90 * time.Second))
				var env envelope
				if err := json.Unmarshal(msg, &env); err != nil {
					continue
				}
				ch, ok := parseEnvelope(env)
				if !ok {
					continue
				}
				select {
				case out <- ch:
				case <-ctx.Done():
					return
				}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
				return
			case <-readDone:
				return
			case <-ticker.C:
				_ = conn.WriteJSON(map[string]any{
					"topic":   "phoenix",
					"event":   "heartbeat",
					"payload": map[string]any{},
					"ref":     fmt.Sprintf("%d", time.Now().UnixNano()),
				})
			}
		}
	}()

	return out, errs, nil
}

type envelope struct {
	Event   string          `json:"event"`
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
}

func joinPayload(accessToken string) map[string]any {
	return map[string]any{
		"config": map[string]any{
			"broadcast": map[string]any{"ack": false, "self": false},
			"presence": map[string]any{
				"enabled": false,
			},
			"postgres_changes": []map[string]string{
				{"event": "*", "schema": "public", "table": "items"},
			},
		},
		"access_token": accessToken,
	}
}

func parseEnvelope(env envelope) (Change, bool) {
	if strings.TrimSpace(env.Event) != "postgres_changes" {
		return Change{}, false
	}
	var p map[string]any
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return Change{}, false
	}
	data, _ := p["data"].(map[string]any)

	// Supabase Realtime payloads appear in both of these shapes:
	// 1) data.eventType + data.new + data.old
	// 2) data.type + data.record + data.old_record
	t := strings.ToUpper(strings.TrimSpace(readString(data, "eventType")))
	if t == "" {
		t = strings.ToUpper(strings.TrimSpace(readString(data, "type")))
	}
	if t == "" {
		return Change{}, false
	}
	n := readMap(data, "new")
	if len(n) == 0 {
		n = readMap(data, "record")
	}
	o := readMap(data, "old")
	if len(o) == 0 {
		o = readMap(data, "old_record")
	}
	return Change{Type: t, New: n, Old: o}, true
}

func readString(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func readMap(m map[string]any, key string) map[string]any {
	v, _ := m[key].(map[string]any)
	return v
}

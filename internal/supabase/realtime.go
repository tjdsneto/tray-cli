package supabase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

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
	if base == "" {
		return nil, fmt.Errorf("supabase realtime: empty base URL")
	}
	if strings.TrimSpace(anonKey) == "" {
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

// SubscribeItems subscribes to public.items changes for the current JWT.
func (c *RealtimeClient) SubscribeItems(ctx context.Context, accessToken string) (<-chan Change, <-chan error, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, nil, fmt.Errorf("supabase realtime: missing access token")
	}
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.wsURL, nil)
	if err != nil {
		return nil, nil, err
	}
	out := make(chan Change, 32)
	errs := make(chan error, 1)

	join := map[string]any{
		"topic":   "realtime:public:items",
		"event":   "phx_join",
		"payload": joinPayload(accessToken),
		"ref":     "1",
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
			"presence":  map[string]any{"key": ""},
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
	var p struct {
		Data struct {
			EventType string         `json:"eventType"`
			New       map[string]any `json:"new"`
			Old       map[string]any `json:"old"`
		} `json:"data"`
	}
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return Change{}, false
	}
	t := strings.ToUpper(strings.TrimSpace(p.Data.EventType))
	if t == "" {
		return Change{}, false
	}
	return Change{Type: t, New: p.Data.New, Old: p.Data.Old}, true
}

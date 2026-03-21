package postgrest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tjdsneto/tray-cli/internal/domain"
	supabasehttp "github.com/tjdsneto/tray-cli/internal/supabase"
)

// client wraps the low-level HTTP client with JSON request/response helpers.
type client struct {
	http *supabasehttp.Client
}

func newClient(c *supabasehttp.Client) *client {
	return &client{http: c}
}

func hdrPreferRepresentation() http.Header {
	h := http.Header{}
	h.Set("Prefer", "return=representation")
	return h
}

func (p *client) request(ctx context.Context, sess domain.Session, method, path string, body any, extra http.Header) ([]byte, error) {
	var rdr io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("postgrest: encode body: %w", err)
		}
		rdr = bytes.NewReader(buf)
	}
	req, err := p.http.NewRequest(ctx, method, path, rdr, sess.AccessToken)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for k, vv := range extra {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	resp, err := p.http.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("postgrest: request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, httpAPIError(method, path, resp.Status, resp.StatusCode, raw)
	}
	return raw, nil
}

func (p *client) doJSON(ctx context.Context, sess domain.Session, method, path string, body any, out any, extra http.Header) error {
	raw, err := p.request(ctx, sess, method, path, body, extra)
	if err != nil {
		return err
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("postgrest: decode response: %w", err)
	}
	return nil
}

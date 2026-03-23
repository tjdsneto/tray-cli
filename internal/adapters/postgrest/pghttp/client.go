// Package pghttp holds generic Supabase REST (PostgREST-style) HTTP helpers.
// Domain-specific adapters in the parent package use this client.
package pghttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	supabasehttp "github.com/tjdsneto/tray-cli/internal/supabase"
)

// Client performs authenticated JSON requests against the Supabase Data API.
type Client struct {
	http *supabasehttp.Client
}

// New wraps a Supabase HTTP client for REST calls.
func New(c *supabasehttp.Client) *Client {
	return &Client{http: c}
}

// PreferRepresentation returns Prefer: return=representation for mutating requests.
func PreferRepresentation() http.Header {
	h := http.Header{}
	h.Set("Prefer", "return=representation")
	return h
}

// Request sends method path with optional JSON body; bearerToken is the user JWT (or anon for some calls).
func (p *Client) Request(ctx context.Context, bearerToken, method, path string, body any, extra http.Header) ([]byte, error) {
	var rdr io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("pghttp: encode body: %w", err)
		}
		rdr = bytes.NewReader(buf)
	}
	req, err := p.http.NewRequest(ctx, method, path, rdr, bearerToken)
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
		return nil, fmt.Errorf("pghttp: request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, apiError(method, path, resp.Status, resp.StatusCode, raw)
	}
	return raw, nil
}

// DoJSON is like Request then json.Unmarshal into out when the body is non-empty.
func (p *Client) DoJSON(ctx context.Context, bearerToken, method, path string, body any, out any, extra http.Header) error {
	raw, err := p.Request(ctx, bearerToken, method, path, body, extra)
	if err != nil {
		return err
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("pghttp: decode response: %w", err)
	}
	return nil
}

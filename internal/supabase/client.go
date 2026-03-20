package supabase

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client is a thin PostgREST-oriented HTTP client (no Supabase SDK).
type Client struct {
	baseURL    *url.URL
	anonKey    string
	HTTPClient *http.Client
}

// NewClient parses rawBaseURL (e.g. https://xyz.supabase.co) and stores anonKey for apikey header.
func NewClient(rawBaseURL, anonKey string, httpClient *http.Client) (*Client, error) {
	if strings.TrimSpace(rawBaseURL) == "" {
		return nil, fmt.Errorf("supabase: empty base URL")
	}
	if strings.TrimSpace(anonKey) == "" {
		return nil, fmt.Errorf("supabase: empty anon key")
	}
	u, err := url.Parse(strings.TrimRight(rawBaseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("supabase: parse URL: %w", err)
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: u, anonKey: anonKey, HTTPClient: httpClient}, nil
}

// NewRequest builds a request to path (e.g. /rest/v1/trays). accessToken is optional (Bearer).
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader, accessToken string) (*http.Request, error) {
	full := strings.TrimRight(c.baseURL.String(), "/") + "/" + strings.TrimLeft(path, "/")
	req, err := http.NewRequestWithContext(ctx, method, full, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", c.anonKey)
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	return req, nil
}

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// User is a subset of Supabase GoTrue GET /auth/v1/user.
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// FetchUser calls GET /auth/v1/user with the access token (and anon apikey header).
func FetchUser(ctx context.Context, projectURL, anonKey, accessToken string, httpClient *http.Client) (User, error) {
	projectURL = strings.TrimRight(strings.TrimSpace(projectURL), "/")
	if projectURL == "" {
		return User{}, fmt.Errorf("auth: empty project URL")
	}
	if strings.TrimSpace(anonKey) == "" {
		return User{}, fmt.Errorf("auth: empty anon key")
	}
	if strings.TrimSpace(accessToken) == "" {
		return User{}, fmt.Errorf("auth: empty access token")
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	url := projectURL + "/auth/v1/user"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return User{}, err
	}
	req.Header.Set("apikey", anonKey)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("auth: request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return User{}, fmt.Errorf("auth: %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}
	var u User
	if err := json.Unmarshal(raw, &u); err != nil {
		return User{}, fmt.Errorf("auth: decode user: %w", err)
	}
	if u.ID == "" {
		return User{}, fmt.Errorf("auth: user response missing id")
	}
	return u, nil
}

// FetchUserTimeout wraps FetchUser with a default timeout.
func FetchUserTimeout(projectURL, anonKey, accessToken string, httpClient *http.Client) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return FetchUser(ctx, projectURL, anonKey, accessToken, httpClient)
}

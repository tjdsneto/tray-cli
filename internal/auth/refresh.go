package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/credentials"
)

const defaultRefreshLeeway = 2 * time.Minute

// RefreshTokens exchanges a Supabase refresh token for new session tokens.
func RefreshTokens(ctx context.Context, projectURL, anonKey, refreshToken string, httpClient *http.Client) (accessToken, newRefreshToken, userID string, err error) {
	base, err := authV1Base(projectURL)
	if err != nil {
		return "", "", "", err
	}
	if strings.TrimSpace(anonKey) == "" {
		return "", "", "", fmt.Errorf("auth: empty anon key")
	}
	if strings.TrimSpace(refreshToken) == "" {
		return "", "", "", fmt.Errorf("auth: empty refresh token")
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	u := base + "/token?grant_type=refresh_token"
	body, err := json.Marshal(map[string]string{"refresh_token": strings.TrimSpace(refreshToken)})
	if err != nil {
		return "", "", "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", anonKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("auth: refresh request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", "", fmt.Errorf("auth: refresh %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}
	var out struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", "", "", fmt.Errorf("auth: decode refresh response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", "", "", fmt.Errorf("auth: refresh response missing access_token")
	}
	uid := strings.TrimSpace(out.User.ID)
	return out.AccessToken, out.RefreshToken, uid, nil
}

// EnsureFreshCredentials returns creds with a valid access token when a refresh token is stored.
// It refreshes (and persists) when the access JWT is missing, unreadable, or expires within leeway.
// Manual login (--token) with no refresh_token is returned unchanged.
func EnsureFreshCredentials(ctx context.Context, projectURL, anonKey string, httpClient *http.Client, configDir string, f credentials.File) (credentials.File, error) {
	if strings.TrimSpace(f.RefreshToken) == "" {
		return f, nil
	}
	if !accessTokenNeedsRefresh(f.AccessToken, defaultRefreshLeeway) {
		return f, nil
	}
	access, refresh, uid, err := RefreshTokens(ctx, projectURL, anonKey, f.RefreshToken, httpClient)
	if err != nil {
		return f, fmt.Errorf("session expired — run `tray login` again (%w)", err)
	}
	out := credentials.File{
		AccessToken:  access,
		RefreshToken: f.RefreshToken,
		UserID:       f.UserID,
	}
	if strings.TrimSpace(refresh) != "" {
		out.RefreshToken = refresh
	}
	if strings.TrimSpace(uid) != "" {
		out.UserID = uid
	}
	if err := credentials.Save(configDir, out); err != nil {
		return f, err
	}
	return out, nil
}

package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// AuthorizeURL builds the Supabase Auth OAuth URL (PKCE / S256).
func AuthorizeURL(projectURL, provider, redirectTo, codeChallenge string) (string, error) {
	base, err := authV1Base(projectURL)
	if err != nil {
		return "", err
	}
	v := url.Values{}
	v.Set("provider", provider)
	v.Set("redirect_to", redirectTo)
	v.Set("code_challenge", codeChallenge)
	v.Set("code_challenge_method", "s256")
	return base + "/authorize?" + v.Encode(), nil
}

func authV1Base(projectURL string) (string, error) {
	u := strings.TrimRight(strings.TrimSpace(projectURL), "/")
	if u == "" {
		return "", fmt.Errorf("auth: empty project URL")
	}
	return u + "/auth/v1", nil
}

// pkceTokenResponse is the JSON body from POST /auth/v1/token?grant_type=pkce.
type pkceTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	User         struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

// ExchangePKCE swaps auth_code + code_verifier for session tokens.
func ExchangePKCE(ctx context.Context, projectURL, anonKey, authCode, codeVerifier string, httpClient *http.Client) (pkceTokenResponse, error) {
	var zero pkceTokenResponse
	base, err := authV1Base(projectURL)
	if err != nil {
		return zero, err
	}
	if strings.TrimSpace(anonKey) == "" {
		return zero, fmt.Errorf("auth: empty anon key")
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	u := base + "/token?grant_type=pkce"
	body := map[string]string{
		"auth_code":     strings.TrimSpace(authCode),
		"code_verifier": codeVerifier,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return zero, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return zero, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", anonKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("auth: token exchange: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(respBody))
		if strings.Contains(msg, "provider is not enabled") {
			msg += "\n\nHint: Supabase → Authentication → Providers → enable Google (or your chosen provider), save Client ID and Secret from Google Cloud, then retry."
		}
		return zero, fmt.Errorf("auth: token exchange %s: %s", resp.Status, msg)
	}
	var out pkceTokenResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return zero, fmt.Errorf("auth: decode token response: %w", err)
	}
	if out.AccessToken == "" {
		return zero, fmt.Errorf("auth: token response missing access_token")
	}
	if out.User.ID == "" {
		return zero, fmt.Errorf("auth: token response missing user id")
	}
	return out, nil
}

// LoginWithOAuth runs the browser OAuth + PKCE flow and returns tokens from Supabase Auth.
// If provider is non-empty, the browser opens that provider’s authorize URL directly.
// If provider is empty, the browser opens a local page listing providers to choose from (same PKCE session).
// If onListening is non-nil, it is called with redirect_to, the picker URL, and the direct provider (if any).
func LoginWithOAuth(ctx context.Context, projectURL, anonKey, provider string, httpClient *http.Client, onListening func(redirectURL string, pickURL string, directProvider string)) (accessToken, refreshToken, userID, email string, err error) {
	verifier, challenge, err := NewCodeVerifier()
	if err != nil {
		return "", "", "", "", err
	}
	redirectTo, pickURL, wait, srv, err := OAuthCallbackServer(projectURL, challenge)
	if err != nil {
		return "", "", "", "", err
	}
	defer func() { _ = srv.Shutdown(context.Background()) }()

	direct := strings.TrimSpace(provider)
	if onListening != nil {
		onListening(redirectTo, pickURL, direct)
	}

	if direct != "" {
		authURL, err := AuthorizeURL(projectURL, direct, redirectTo, challenge)
		if err != nil {
			return "", "", "", "", err
		}
		if err := OpenURL(authURL); err != nil {
			return "", "", "", "", fmt.Errorf("auth: open browser: %w", err)
		}
	} else {
		if err := OpenURL(pickURL); err != nil {
			return "", "", "", "", fmt.Errorf("auth: open browser: %w", err)
		}
	}

	waitCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	code, cbErr := wait(waitCtx)
	if cbErr != nil {
		return "", "", "", "", cbErr
	}
	if code == "" {
		return "", "", "", "", fmt.Errorf("auth: empty auth code")
	}
	tok, err := ExchangePKCE(waitCtx, projectURL, anonKey, code, verifier, httpClient)
	if err != nil {
		return "", "", "", "", err
	}
	return tok.AccessToken, tok.RefreshToken, tok.User.ID, tok.User.Email, nil
}

// OAuthCallbackServer listens on 127.0.0.1:0, serves a provider picker at / and callback at /callback.
func OAuthCallbackServer(projectURL, codeChallenge string) (redirectTo string, pickURL string, wait func(context.Context) (code string, oauthErr error), srv *http.Server, err error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("auth: listen: %w", err)
	}
	addr := ln.Addr().(*net.TCPAddr)
	redirect := fmt.Sprintf("http://127.0.0.1:%d/callback", addr.Port)
	picker := fmt.Sprintf("http://127.0.0.1:%d/", addr.Port)

	links, err := buildPickerLinks(projectURL, redirect, codeChallenge)
	if err != nil {
		_ = ln.Close()
		return "", "", nil, nil, err
	}

	ch := make(chan struct {
		code string
		err  error
	}, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		html, err := renderPickerPage(links)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, html)
	})
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		q := r.URL.Query()
		if e := q.Get("error"); e != "" {
			desc := q.Get("error_description")
			ch <- struct {
				code string
				err  error
			}{"", fmt.Errorf("%s: %s", e, desc)}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = io.WriteString(w, oauthDoneHTML(false))
			return
		}
		code := q.Get("code")
		if code == "" {
			ch <- struct {
				code string
				err  error
			}{"", fmt.Errorf("auth: missing code in callback (check redirect URL allow list in Supabase)")}
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}
		ch <- struct {
			code string
			err  error
		}{code, nil}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, oauthDoneHTML(true))
	})

	srv = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = srv.Serve(ln) }()

	waitFn := func(waitCtx context.Context) (string, error) {
		select {
		case <-waitCtx.Done():
			return "", waitCtx.Err()
		case res := <-ch:
			return res.code, res.err
		}
	}

	return redirect, picker, waitFn, srv, nil
}

func oauthDoneHTML(ok bool) string {
	if ok {
		return `<!doctype html><meta charset="utf-8"><title>Tray</title><p>Sign-in complete. You can close this tab and return to the terminal.</p>`
	}
	return `<!doctype html><meta charset="utf-8"><title>Tray</title><p>Sign-in failed. You can close this tab and check the terminal.</p>`
}

// OpenURL opens u in the default browser (best effort).
func OpenURL(u string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", u).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
	default:
		return exec.Command("xdg-open", u).Start()
	}
}

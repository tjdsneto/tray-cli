package postgrest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/config"
)

type supabaseErrJSON struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func httpAPIError(method, path, statusLine string, statusCode int, raw []byte) error {
	rawStr := strings.TrimSpace(string(raw))
	if config.Debug() {
		return fmt.Errorf("postgrest: %s %s: %s: %s", method, path, statusLine, rawStr)
	}

	var j supabaseErrJSON
	_ = json.Unmarshal(raw, &j)
	code := strings.TrimSpace(j.Code)
	msg := sanitizeOneLine(strings.TrimSpace(j.Message))

	switch statusCode {
	case http.StatusBadRequest: // 400
		if msg != "" {
			return fmt.Errorf("invalid request: %s", msg)
		}
		return fmt.Errorf("invalid request (%s)", statusLine)
	case http.StatusUnauthorized: // 401
		return fmt.Errorf("session expired or invalid; run tray login")
	case http.StatusForbidden: // 403
		if msg != "" {
			return fmt.Errorf("not allowed: %s", msg)
		}
		return fmt.Errorf("not allowed")
	case http.StatusConflict: // 409 — unique violations, etc.
		if code == "23505" || strings.Contains(strings.ToLower(msg), "duplicate") {
			return fmt.Errorf("that name is already in use (pick another)")
		}
		if msg != "" {
			return fmt.Errorf("conflict: %s", msg)
		}
		return fmt.Errorf("request conflict (%s)", statusLine)
	case http.StatusNotFound: // 404
		if msg != "" {
			return fmt.Errorf("not found: %s", msg)
		}
		return fmt.Errorf("not found")
	case http.StatusUnprocessableEntity: // 422
		if msg != "" {
			return fmt.Errorf("invalid data: %s", msg)
		}
		return fmt.Errorf("invalid data (%s)", statusLine)
	case http.StatusInternalServerError:
		switch code {
		case "42P17":
			return fmt.Errorf(
				"server access rules failed (database policy). Apply the latest Supabase migration from this repo (e.g. supabase db push) or ask a project admin",
			)
		case "42501":
			return fmt.Errorf("permission denied")
		}
		if msg != "" {
			return fmt.Errorf("server error: %s", msg)
		}
		return fmt.Errorf("server error (%s)", statusLine)
	default:
		if msg != "" {
			return fmt.Errorf("request failed: %s", msg)
		}
		return fmt.Errorf("request failed (%s)", statusLine)
	}
}

func sanitizeOneLine(s string) string {
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "\n", " ")
	const max = 220
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}

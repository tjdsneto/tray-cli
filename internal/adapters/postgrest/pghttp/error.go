package pghttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/config"
)

// ErrUnauthorized is returned (wrapped) when the API responds with HTTP 401.
// Use errors.Is(err, ErrUnauthorized) to detect an expired or rejected access token.
var ErrUnauthorized = errors.New("pghttp: unauthorized")

type supabaseErrJSON struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func apiError(method, path, statusLine string, statusCode int, raw []byte) error {
	rawStr := strings.TrimSpace(string(raw))
	if config.Debug() {
		if statusCode == http.StatusUnauthorized {
			return fmt.Errorf("%w: pghttp: %s %s: %s: %s", ErrUnauthorized, method, path, statusLine, rawStr)
		}
		return fmt.Errorf("pghttp: %s %s: %s: %s", method, path, statusLine, rawStr)
	}

	var j supabaseErrJSON
	_ = json.Unmarshal(raw, &j)
	code := strings.TrimSpace(j.Code)
	msg := sanitizeOneLine(strings.TrimSpace(j.Message))
	msgLower := strings.ToLower(msg)

	switch statusCode {
	case http.StatusBadRequest:
		if msg != "" {
			return fmt.Errorf("that request wasn't valid: %s", msg)
		}
		return fmt.Errorf("that request wasn't valid (%s)", statusLine)
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: your session expired or isn't valid — run `tray login` and try again", ErrUnauthorized)
	case http.StatusForbidden:
		if msg != "" {
			return fmt.Errorf("you're not allowed to do that: %s", msg)
		}
		return fmt.Errorf("you're not allowed to do that (permission denied)")
	case http.StatusConflict:
		if isDuplicateConflict(code, msgLower) {
			return duplicateConflictMessage(path)
		}
		if msg != "" {
			return fmt.Errorf("that conflicts with existing data: %s", msg)
		}
		return fmt.Errorf("that conflicts with existing data (%s)", statusLine)
	case http.StatusNotFound:
		if msg != "" {
			return fmt.Errorf("nothing matched that request: %s", msg)
		}
		return fmt.Errorf("nothing matched that request (not found)")
	case http.StatusUnprocessableEntity:
		if msg != "" {
			return fmt.Errorf("that data couldn't be saved: %s", msg)
		}
		return fmt.Errorf("that data couldn't be saved (%s)", statusLine)
	case http.StatusInternalServerError:
		if isDuplicateConflict(code, msgLower) {
			return duplicateConflictMessage(path)
		}
		switch code {
		case "42P17":
			return fmt.Errorf(
				"the server blocked this action (database access rules). Ask your admin to apply the latest migrations from this repo, or run supabase db push if you manage the project",
			)
		case "42501":
			return fmt.Errorf("permission denied on the server")
		}
		if msg != "" {
			return fmt.Errorf("something went wrong on the server: %s", msg)
		}
		return fmt.Errorf("something went wrong on the server (%s)", statusLine)
	default:
		if msg != "" {
			return fmt.Errorf("request didn't succeed: %s", msg)
		}
		return fmt.Errorf("request didn't succeed (%s)", statusLine)
	}
}

func isDuplicateConflict(pgCode, msgLower string) bool {
	if pgCode == "23505" {
		return true
	}
	return strings.Contains(msgLower, "duplicate") || strings.Contains(msgLower, "unique constraint")
}

func duplicateConflictMessage(path string) error {
	switch {
	case strings.Contains(path, "/rest/v1/trays"):
		return fmt.Errorf("you already have a tray with that name — choose another name, or run `tray ls` to see your trays")
	default:
		return fmt.Errorf("that value is already taken — try a different name or pick another option")
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

package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/bar"
)

type ErrorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type Session struct {
	SignedIn bool `json:"signedIn"`
}

type Snapshot struct {
	Badge        int       `json:"badge"`
	RemoteStatus string    `json:"remoteStatus"`
	Sections     []Section `json:"sections"`
}

type Section struct {
	Header string `json:"header"`
	Rows   []Row  `json:"rows"`
}

type Row struct {
	Scope     string `json:"scope"`
	TrayRef   string `json:"trayRef"`
	ItemID    string `json:"itemId"`
	Title     string `json:"title"`
	Clipboard string `json:"clipboard"`
}

type Tray struct {
	Scope     string `json:"scope"`
	Ref       string `json:"ref"`
	Name      string `json:"name"`
	OpenCount int    `json:"openCount"`
}

type Item struct {
	ID          string     `json:"id"`
	Scope       string     `json:"scope"`
	TrayRef     string     `json:"trayRef"`
	Title       string     `json:"title"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	SnoozeUntil *time.Time `json:"snoozeUntil,omitempty"`
	Body        *string    `json:"body,omitempty"`
	Actions     []string   `json:"actions"`
}

type AddRequest struct {
	Tray  string  `json:"tray"`
	Title string  `json:"title"`
	Body  *string `json:"body,omitempty"`
}

type DeclineRequest struct {
	Reason string `json:"reason"`
}

type SnoozeRequest struct {
	Until string `json:"until"`
}

type invalidError string

func (e invalidError) Error() string { return string(e) }

func errInvalid(msg string) error { return invalidError(msg) }

func Invalid(msg string) error { return invalidError(msg) }

func SnapshotFrom(s bar.Snapshot) Snapshot {
	out := Snapshot{
		Badge:        s.Badge,
		RemoteStatus: s.RemoteStatus,
		Sections:     []Section{},
	}
	if len(s.Sections) == 0 {
		return out
	}
	out.Sections = make([]Section, 0, len(s.Sections))
	for _, sec := range s.Sections {
		apiSec := Section{Header: sec.Header}
		for _, row := range sec.Rows {
			apiSec.Rows = append(apiSec.Rows, Row{
				Scope:     string(row.Scope),
				TrayRef:   row.TrayRef,
				ItemID:    row.ItemID,
				Title:     row.Title,
				Clipboard: bar.ClipboardLine(row),
			})
		}
		out.Sections = append(out.Sections, apiSec)
	}
	return out
}

func ValidateAdd(r AddRequest) error {
	if strings.TrimSpace(r.Tray) == "" {
		return errInvalid("tray is required")
	}
	if strings.TrimSpace(r.Title) == "" {
		return errInvalid("title is required")
	}
	return nil
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorBody{Error: code, Message: message})
}

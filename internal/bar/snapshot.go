package bar

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Scope string

const (
	ScopeLocal  Scope = "local"
	ScopeRemote Scope = "remote"
)

// Item is one attention-backlog row before grouping.
type Item struct {
	Scope     Scope
	TrayRef   string // first clipboard token
	TrayName  string // group key (display name before disambiguation)
	ItemID    string
	Title     string
	CreatedAt time.Time
}

type Row struct {
	Scope   Scope
	TrayRef string
	ItemID  string
	Title   string
}

type Section struct {
	Header string
	Rows   []Row
}

type Snapshot struct {
	Badge        int
	Sections     []Section
	RemoteStatus string // muted status line; empty if OK
}

// ClipboardLine formats tray-ref and item-id for paste into an agent chat.
// If TrayRef contains whitespace, it is double-quoted so the line stays two fields.
func ClipboardLine(r Row) string {
	ref := r.TrayRef
	if strings.ContainsAny(ref, " \t") {
		ref = strconv.Quote(ref)
	}
	return fmt.Sprintf("%s %s", ref, r.ItemID)
}

// ParseClipboardLine splits a clipboard handoff into tray-ref and item-id.
// Accepts unquoted (`inbox abc`) and quoted (`"dir:/path with spaces" abc`) forms.
func ParseClipboardLine(s string) (trayRef, itemID string, err error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", "", fmt.Errorf("empty clipboard line")
	}
	if s[0] == '"' {
		end := 1
		for end < len(s) {
			if s[end] == '\\' && end+1 < len(s) {
				end += 2
				continue
			}
			if s[end] == '"' {
				break
			}
			end++
		}
		if end >= len(s) || s[end] != '"' {
			return "", "", fmt.Errorf("unclosed quote in clipboard line")
		}
		trayRef, err = strconv.Unquote(s[:end+1])
		if err != nil {
			return "", "", fmt.Errorf("invalid quoted tray-ref: %w", err)
		}
		rest := strings.TrimSpace(s[end+1:])
		if rest == "" {
			return "", "", fmt.Errorf("missing item-id after tray-ref")
		}
		if strings.ContainsAny(rest, " \t") {
			return "", "", fmt.Errorf("expected single item-id after tray-ref")
		}
		return trayRef, rest, nil
	}
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected tray-ref and item-id, got %d tokens", len(parts))
	}
	return parts[0], parts[1], nil
}

// RemoteItemsFromSnapshot collects ScopeRemote rows as Items for a degraded refresh
// (fresh local + last-known remote). TrayName defaults to TrayRef; CreatedAt is zero.
func RemoteItemsFromSnapshot(snap Snapshot) []Item {
	var out []Item
	for _, sec := range snap.Sections {
		for _, row := range sec.Rows {
			if row.Scope != ScopeRemote {
				continue
			}
			out = append(out, Item{
				Scope: ScopeRemote, TrayRef: row.TrayRef, TrayName: row.TrayRef,
				ItemID: row.ItemID, Title: row.Title,
			})
		}
	}
	return out
}

// BuildSnapshot groups items by tray name (A–Z), newest-first within a section.
// When the same TrayName appears in both scopes, headers become "name · local" / "name · remote".
func BuildSnapshot(items []Item, remoteStatus string) Snapshot {
	type key struct {
		name  string
		scope Scope
	}
	nameScopes := map[string]map[Scope]struct{}{}
	for _, it := range items {
		if nameScopes[it.TrayName] == nil {
			nameScopes[it.TrayName] = map[Scope]struct{}{}
		}
		nameScopes[it.TrayName][it.Scope] = struct{}{}
	}
	buckets := map[key][]Item{}
	for _, it := range items {
		k := key{name: it.TrayName, scope: it.Scope}
		buckets[k] = append(buckets[k], it)
	}
	var keys []key
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].name != keys[j].name {
			return keys[i].name < keys[j].name
		}
		return keys[i].scope < keys[j].scope
	})
	snap := Snapshot{Badge: len(items), RemoteStatus: remoteStatus}
	for _, k := range keys {
		list := buckets[k]
		sort.Slice(list, func(i, j int) bool {
			return list[i].CreatedAt.After(list[j].CreatedAt)
		})
		header := k.name
		if len(nameScopes[k.name]) > 1 {
			header = fmt.Sprintf("%s · %s", k.name, k.scope)
		}
		sec := Section{Header: header}
		for _, it := range list {
			sec.Rows = append(sec.Rows, Row{
				Scope: it.Scope, TrayRef: it.TrayRef, ItemID: it.ItemID, Title: it.Title,
			})
		}
		snap.Sections = append(snap.Sections, sec)
	}
	return snap
}

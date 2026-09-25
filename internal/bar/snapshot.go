package bar

import (
	"fmt"
	"sort"
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

func ClipboardLine(r Row) string {
	return fmt.Sprintf("%s %s", r.TrayRef, r.ItemID)
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

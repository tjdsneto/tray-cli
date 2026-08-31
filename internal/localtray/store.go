package localtray

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	StatusOpen = "open"
	StatusDone = "done"
)

// Item is a line on a local tray.
type Item struct {
	ID        string     `json:"id"`
	TrayID    string     `json:"tray_id"`
	Title     string     `json:"title"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	DoneAt    *time.Time `json:"done_at,omitempty"`
}

// ItemWithTray pairs an item with its tray record for listing.
type ItemWithTray struct {
	Item Item
	Tray TrayRecord
}

// Store manages the local tray index and item files under configDir/local.
type Store struct {
	ConfigDir string
}

func NewStore(configDir string) *Store {
	return &Store{ConfigDir: configDir}
}

func (s *Store) localDir() string {
	return LocalDir(s.ConfigDir)
}

func (s *Store) loadIndex() (Index, error) {
	return LoadIndex(s.localDir())
}

func (s *Store) saveIndex(idx Index) error {
	return SaveIndex(s.localDir(), idx)
}

func itemsFilePath(localDir, rel string) string {
	return filepath.Join(localDir, filepath.FromSlash(rel))
}

func newItemID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func hashItemsFileName(trayID string) string {
	sum := sha256.Sum256([]byte(trayID))
	return hex.EncodeToString(sum[:8]) + ".jsonl"
}

// EnsureFromTarget registers a tray for target if missing and returns the record.
func (s *Store) EnsureFromTarget(t AddTarget) (TrayRecord, error) {
	switch t.Kind {
	case KindGlobal:
		id, err := GlobalID(t.Name)
		if err != nil {
			return TrayRecord{}, err
		}
		return s.ensure(id, func() TrayRecord {
			return TrayRecord{
				ID:   id,
				Kind: KindGlobal,
				Name: normalizeGlobalName(t.Name),
			}
		})
	case KindDir:
		id, err := DirID(t.Path)
		if err != nil {
			return TrayRecord{}, err
		}
		abs, _ := filepath.Abs(t.Path)
		return s.ensure(id, func() TrayRecord {
			return TrayRecord{
				ID:   id,
				Kind: KindDir,
				Path: filepath.Clean(abs),
			}
		})
	case KindBranch:
		id, err := BranchID(t.RepoRoot, t.Branch)
		if err != nil {
			return TrayRecord{}, err
		}
		root, _ := filepath.Abs(t.RepoRoot)
		return s.ensure(id, func() TrayRecord {
			return TrayRecord{
				ID:       id,
				Kind:     KindBranch,
				RepoRoot: filepath.Clean(root),
				Branch:   strings.TrimSpace(t.Branch),
			}
		})
	default:
		return TrayRecord{}, fmt.Errorf("unknown tray kind %q", t.Kind)
	}
}

func (s *Store) ensure(id string, build func() TrayRecord) (TrayRecord, error) {
	idx, err := s.loadIndex()
	if err != nil {
		return TrayRecord{}, err
	}
	if rec, ok := idx.Find(id); ok {
		return rec, nil
	}
	rec := build()
	rec.CreatedAt = time.Now().UTC()
	rec.ItemsFile = filepath.ToSlash(filepath.Join("items", hashItemsFileName(id)))
	idx.Trays = append(idx.Trays, rec)
	if err := s.saveIndex(idx); err != nil {
		return TrayRecord{}, err
	}
	localDir := s.localDir()
	if err := os.MkdirAll(ItemsDir(localDir), 0o700); err != nil {
		return TrayRecord{}, err
	}
	p := itemsFilePath(localDir, rec.ItemsFile)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return TrayRecord{}, err
	}
	_ = f.Close()
	return rec, nil
}

// AddItem appends an open item to the tray (creates tray via Ensure when needed).
func (s *Store) AddItem(target AddTarget, title string, create bool) (Item, TrayRecord, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Item{}, TrayRecord{}, fmt.Errorf("item title cannot be empty")
	}
	id, err := trayIDFromTarget(target)
	if err != nil {
		return Item{}, TrayRecord{}, err
	}
	idx, err := s.loadIndex()
	if err != nil {
		return Item{}, TrayRecord{}, err
	}
	rec, ok := idx.Find(id)
	if !ok {
		if !create {
			return Item{}, TrayRecord{}, fmt.Errorf("local tray %q does not exist — omit --no-create to create it with the item", DisplayLabel(id))
		}
		rec, err = s.EnsureFromTarget(target)
		if err != nil {
			return Item{}, TrayRecord{}, err
		}
	}
	itemID, err := newItemID()
	if err != nil {
		return Item{}, TrayRecord{}, err
	}
	item := Item{
		ID:        itemID,
		TrayID:    rec.ID,
		Title:     title,
		Status:    StatusOpen,
		CreatedAt: time.Now().UTC(),
	}
	if err := appendItem(s.localDir(), rec, item); err != nil {
		return Item{}, TrayRecord{}, err
	}
	return item, rec, nil
}

func trayIDFromTarget(t AddTarget) (string, error) {
	switch t.Kind {
	case KindGlobal:
		return GlobalID(t.Name)
	case KindDir:
		return DirID(t.Path)
	case KindBranch:
		return BranchID(t.RepoRoot, t.Branch)
	default:
		return "", fmt.Errorf("unknown tray kind %q", t.Kind)
	}
}

func readItems(localDir string, rec TrayRecord) ([]Item, error) {
	p := itemsFilePath(localDir, rec.ItemsFile)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var items []Item
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var it Item
		if err := json.Unmarshal(line, &it); err != nil {
			return nil, fmt.Errorf("parse items file %s: %w", rec.ItemsFile, err)
		}
		items = append(items, it)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func appendItem(localDir string, rec TrayRecord, item Item) error {
	line, err := json.Marshal(item)
	if err != nil {
		return err
	}
	p := itemsFilePath(localDir, rec.ItemsFile)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		return err
	}
	return nil
}

func writeItems(localDir string, rec TrayRecord, items []Item) error {
	p := itemsFilePath(localDir, rec.ItemsFile)
	var buf bytes.Buffer
	for _, it := range items {
		line, err := json.Marshal(it)
		if err != nil {
			return err
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// ListItemsForTrays returns open items for the given tray ids.
func (s *Store) ListItemsForTrays(trayIDs []string) ([]ItemWithTray, error) {
	idx, err := s.loadIndex()
	if err != nil {
		return nil, err
	}
	localDir := s.localDir()
	var out []ItemWithTray
	for _, tid := range trayIDs {
		rec, ok := idx.Find(tid)
		if !ok {
			continue
		}
		items, err := readItems(localDir, rec)
		if err != nil {
			return nil, err
		}
		for _, it := range items {
			if strings.EqualFold(it.Status, StatusDone) {
				continue
			}
			out = append(out, ItemWithTray{Item: it, Tray: rec})
		}
	}
	return out, nil
}

// TraySummaries returns tray records with live open-item counts.
func (s *Store) TraySummaries(trayIDs []string) ([]TraySummary, error) {
	idx, err := s.loadIndex()
	if err != nil {
		return nil, err
	}
	localDir := s.localDir()
	var out []TraySummary
	for _, tid := range trayIDs {
		rec, ok := idx.Find(tid)
		if !ok {
			continue
		}
		items, err := readItems(localDir, rec)
		if err != nil {
			return nil, err
		}
		open := 0
		for _, it := range items {
			if !strings.EqualFold(it.Status, StatusDone) {
				open++
			}
		}
		out = append(out, TraySummary{Record: rec, OpenCount: open, TotalCount: len(items)})
	}
	return out, nil
}

// TraySummary is a tray plus item counts for ls output.
type TraySummary struct {
	Record     TrayRecord
	OpenCount  int
	TotalCount int
}

// AllOpenItems returns every open local item across the index (for id resolution).
func (s *Store) AllOpenItems() ([]ItemWithTray, error) {
	idx, err := s.loadIndex()
	if err != nil {
		return nil, err
	}
	return s.ListItemsForTrays(AllTrayIDs(idx))
}

// CompleteItem marks a local item done.
func (s *Store) CompleteItem(itemID string) (Item, error) {
	return s.updateItemStatus(itemID, StatusDone)
}

// RemoveItem deletes a local item line from its tray file.
func (s *Store) RemoveItem(itemID string) error {
	idx, err := s.loadIndex()
	if err != nil {
		return err
	}
	localDir := s.localDir()
	for i := range idx.Trays {
		rec := idx.Trays[i]
		items, err := readItems(localDir, rec)
		if err != nil {
			return err
		}
		changed := false
		var kept []Item
		for _, it := range items {
			if it.ID == itemID {
				changed = true
				continue
			}
			kept = append(kept, it)
		}
		if changed {
			return writeItems(localDir, rec, kept)
		}
	}
	return fmt.Errorf("no local item with id %q", itemID)
}

func (s *Store) updateItemStatus(itemID, status string) (Item, error) {
	idx, err := s.loadIndex()
	if err != nil {
		return Item{}, err
	}
	localDir := s.localDir()
	now := time.Now().UTC()
	for i := range idx.Trays {
		rec := idx.Trays[i]
		items, err := readItems(localDir, rec)
		if err != nil {
			return Item{}, err
		}
		changed := false
		var found Item
		for j := range items {
			if items[j].ID != itemID {
				continue
			}
			items[j].Status = status
			if status == StatusDone {
				items[j].DoneAt = &now
			}
			found = items[j]
			changed = true
			break
		}
		if changed {
			if err := writeItems(localDir, rec, items); err != nil {
				return Item{}, err
			}
			return found, nil
		}
	}
	return Item{}, fmt.Errorf("no local item with id %q", itemID)
}

// PruneEmpty removes non-global trays with zero items from the index and deletes empty item files.
func (s *Store) PruneEmpty(dryRun bool) ([]string, error) {
	idx, err := s.loadIndex()
	if err != nil {
		return nil, err
	}
	localDir := s.localDir()
	var pruned []string
	var kept []TrayRecord
	for _, rec := range idx.Trays {
		if rec.Kind == KindGlobal {
			kept = append(kept, rec)
			continue
		}
		items, err := readItems(localDir, rec)
		if err != nil {
			return nil, err
		}
		if len(items) > 0 {
			kept = append(kept, rec)
			continue
		}
		pruned = append(pruned, rec.ID)
		if dryRun {
			continue
		}
		p := itemsFilePath(localDir, rec.ItemsFile)
		_ = os.Remove(p)
	}
	if dryRun || len(pruned) == 0 {
		return pruned, nil
	}
	idx.Trays = kept
	if err := s.saveIndex(idx); err != nil {
		return nil, err
	}
	return pruned, nil
}

// ResolveItemPrefix finds a unique open local item by hex prefix (min 8 chars).
func (s *Store) ResolveItemPrefix(prefix string) (string, error) {
	prefix = strings.TrimSpace(strings.ToLower(prefix))
	if len(prefix) < 8 {
		return "", fmt.Errorf("local item prefix must be at least 8 hex characters")
	}
	items, err := s.AllOpenItems()
	if err != nil {
		return "", err
	}
	var matches []string
	for _, row := range items {
		id := strings.ToLower(row.Item.ID)
		if strings.HasPrefix(id, prefix) {
			matches = append(matches, row.Item.ID)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no local item matches prefix %q", prefix)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("prefix %q matches multiple local items — use a longer prefix or full id", prefix)
	}
}

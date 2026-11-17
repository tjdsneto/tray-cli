package commands

import (
	"context"
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/api"
	"github.com/tjdsneto/tray-cli/internal/bar"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/localtray"
)

type apiSource struct {
	configDir func() string
	auth      func() (domain.Services, domain.Session, error)
}

func newAPISource(configDir func() string, auth func() (domain.Services, domain.Session, error)) *apiSource {
	return &apiSource{configDir: configDir, auth: auth}
}

func (s *apiSource) SignedIn(context.Context) (bool, error) {
	_, _, err := s.auth()
	return err == nil, nil
}

func (s *apiSource) Attention(ctx context.Context) ([]bar.Item, []bar.Item, error) {
	rows, err := localtray.NewStore(s.configDir()).AllOpenItems()
	if err != nil {
		return nil, nil, err
	}
	local := bar.ItemsFromLocal(rows)

	svcs, sess, err := s.auth()
	if err != nil {
		return local, nil, nil
	}
	owned, err := svcs.Trays.ListOwned(ctx, sess)
	if err != nil {
		return local, nil, api.ErrRemote
	}
	items, err := svcs.Items.List(ctx, sess, pendingOwnedQuery(owned))
	if err != nil {
		return local, nil, api.ErrRemote
	}
	return local, bar.ItemsFromRemote(items, trayref.TrayNameMap(owned)), nil
}

func (s *apiSource) Trays(ctx context.Context) ([]api.Tray, error) {
	local, err := s.localTrays()
	if err != nil {
		return nil, err
	}
	svcs, sess, err := s.auth()
	if err != nil {
		return local, nil
	}
	remote, err := s.remoteTrays(ctx, svcs, sess)
	if err != nil {
		return local, err
	}
	return append(local, remote...), nil
}

func (s *apiSource) Items(ctx context.Context, trayRef string) ([]api.Item, error) {
	rec, ok, err := s.findLocal(trayRef)
	if err != nil {
		return nil, err
	}
	if ok {
		rows, err := localtray.NewStore(s.configDir()).ListItemsForTrays([]string{rec.ID})
		if err != nil {
			return nil, err
		}
		out := make([]api.Item, 0, len(rows))
		for _, row := range rows {
			out = append(out, fromLocal(row.Item, row.Tray))
		}
		return out, nil
	}
	return s.remoteItems(ctx, trayRef)
}

func (s *apiSource) Item(ctx context.Context, id string) (api.Item, error) {
	if strings.TrimSpace(id) == "" {
		return api.Item{}, api.ErrNotFound
	}
	row, ok, err := s.findLocalItem(id)
	if err != nil {
		return api.Item{}, err
	}
	if ok {
		return fromLocal(row.Item, row.Tray), nil
	}
	svcs, sess, err := s.auth()
	if err != nil {
		return api.Item{}, api.ErrSignedOut
	}
	items, err := svcs.Items.List(ctx, sess, domain.ListItemsQuery{ItemID: id})
	if err != nil {
		return api.Item{}, api.ErrRemote
	}
	if len(items) == 0 {
		return api.Item{}, api.ErrNotFound
	}
	it := items[0]
	ref := strings.TrimSpace(it.TrayID)
	if owned, oerr := svcs.Trays.ListOwned(ctx, sess); oerr == nil {
		if name := strings.TrimSpace(trayref.TrayNameMap(owned)[it.TrayID]); name != "" {
			ref = name
		}
	}
	return fromRemote(it, ref), nil
}

func (s *apiSource) Add(ctx context.Context, req api.AddRequest) (api.Item, error) {
	if err := api.ValidateAdd(req); err != nil {
		return api.Item{}, err
	}
	target := localtray.AddTarget{Kind: localtray.KindGlobal, Name: strings.TrimSpace(req.Tray)}
	if rec, ok, err := s.findLocal(req.Tray); err != nil {
		return api.Item{}, err
	} else if ok {
		target = localtray.AddTarget{
			Kind:     rec.Kind,
			Name:     rec.Name,
			Path:     rec.Path,
			RepoRoot: rec.RepoRoot,
			Branch:   rec.Branch,
		}
	} else if item, handled, err := s.addOwned(ctx, req); handled || err != nil {
		return item, err
	}
	item, rec, err := localtray.NewStore(s.configDir()).AddItem(target, strings.TrimSpace(req.Title), true, req.Body)
	if err != nil {
		return api.Item{}, err
	}
	return fromLocal(item, rec), nil
}

// addOwned writes onto an owned remote tray when req.Tray is its id or a unique name.
// handled is false when the caller should create a local tray instead.
func (s *apiSource) addOwned(ctx context.Context, req api.AddRequest) (api.Item, bool, error) {
	svcs, sess, err := s.auth()
	if err != nil {
		return api.Item{}, false, nil
	}
	owned, err := svcs.Trays.ListOwned(ctx, sess)
	if err != nil {
		return api.Item{}, true, api.ErrRemote
	}
	tray, err := resolveOwnedTray(owned, req.Tray)
	if err != nil {
		return api.Item{}, false, nil
	}
	created, err := svcs.Items.Add(ctx, sess, tray.ID, strings.TrimSpace(req.Title), nil, nil, req.Body)
	if err != nil || created == nil {
		return api.Item{}, true, api.ErrRemote
	}
	ref := strings.TrimSpace(tray.Name)
	if ref == "" {
		ref = tray.ID
	}
	return fromRemote(*created, ref), true, nil
}

func (s *apiSource) Act(ctx context.Context, id, action string, body api.ActBody) error {
	_, ok, err := s.findLocalItem(id)
	if err != nil {
		return err
	}
	if ok {
		if action != "complete" {
			return api.Invalid("local items only support complete")
		}
		_, err := localtray.NewStore(s.configDir()).CompleteItem(id)
		return err
	}
	svcs, sess, err := s.auth()
	if err != nil {
		return api.ErrSignedOut
	}
	return remoteAct(ctx, svcs, sess, id, action, body)
}

func (s *apiSource) localTrays() ([]api.Tray, error) {
	idx, err := localtray.LoadIndex(localtray.LocalDir(s.configDir()))
	if err != nil {
		return nil, err
	}
	sums, err := localtray.NewStore(s.configDir()).TraySummaries(localtray.AllTrayIDs(idx))
	if err != nil {
		return nil, err
	}
	out := make([]api.Tray, 0, len(sums))
	for _, sum := range sums {
		name := strings.TrimSpace(sum.Record.Name)
		if name == "" {
			name = localtray.DisplayLabel(sum.Record.ID)
		}
		out = append(out, api.Tray{
			Scope:     "local",
			Ref:       bar.LocalTrayRef(sum.Record),
			Name:      name,
			OpenCount: sum.OpenCount,
		})
	}
	return out, nil
}

func (s *apiSource) remoteTrays(ctx context.Context, svcs domain.Services, sess domain.Session) ([]api.Tray, error) {
	owned, err := svcs.Trays.ListOwned(ctx, sess)
	if err != nil {
		return nil, api.ErrRemote
	}
	counts := map[string]int{}
	if ids := ownedTrayIDList(owned); len(ids) > 0 {
		items, err := svcs.Items.List(ctx, sess, domain.ListItemsQuery{TrayIDIn: ids})
		if err != nil {
			return nil, api.ErrRemote
		}
		for _, it := range items {
			if remoteOpenStatus(it.Status) {
				counts[it.TrayID]++
			}
		}
	}
	out := make([]api.Tray, 0, len(owned))
	for _, t := range owned {
		name := strings.TrimSpace(t.Name)
		if name == "" {
			name = t.ID
		}
		out = append(out, api.Tray{
			Scope:     "remote",
			Ref:       t.ID,
			Name:      name,
			OpenCount: counts[t.ID],
		})
	}
	return out, nil
}

func (s *apiSource) remoteItems(ctx context.Context, trayRef string) ([]api.Item, error) {
	svcs, sess, err := s.auth()
	if err != nil {
		return nil, api.ErrSignedOut
	}
	owned, err := svcs.Trays.ListOwned(ctx, sess)
	if err != nil {
		return nil, api.ErrRemote
	}
	tray, err := resolveOwnedTray(owned, trayRef)
	if err != nil {
		return nil, err
	}
	items, err := svcs.Items.List(ctx, sess, domain.ListItemsQuery{TrayID: tray.ID})
	if err != nil {
		return nil, api.ErrRemote
	}
	name := strings.TrimSpace(tray.Name)
	if name == "" {
		name = tray.ID
	}
	out := make([]api.Item, 0, len(items))
	for _, it := range items {
		out = append(out, fromRemote(it, name))
	}
	return out, nil
}

func remoteAct(ctx context.Context, svcs domain.Services, sess domain.Session, id, action string, body api.ActBody) error {
	patch := domain.ItemPatch{}
	switch action {
	case "accept":
		patch.Status = strPtr("accepted")
	case "decline":
		patch.Status = strPtr("declined")
		if reason := strings.TrimSpace(body.Reason); reason != "" {
			patch.DeclineReason = &reason
		}
	case "complete":
		patch.Status = strPtr("completed")
	case "archive":
		patch.Status = strPtr("archived")
	case "snooze":
		until, err := parseSnoozeUntil(body.Until)
		if err != nil {
			return err
		}
		patch.Status = strPtr("snoozed")
		patch.SnoozeUntil = &until
	default:
		return api.Invalid("unknown action")
	}
	return svcs.Items.Update(ctx, sess, id, patch)
}

func (s *apiSource) findLocal(ref string) (localtray.TrayRecord, bool, error) {
	idx, err := localtray.LoadIndex(localtray.LocalDir(s.configDir()))
	if err != nil {
		return localtray.TrayRecord{}, false, err
	}
	want := strings.TrimSpace(ref)
	for i := range idx.Trays {
		if bar.LocalTrayRef(idx.Trays[i]) == want {
			return idx.Trays[i], true, nil
		}
	}
	return localtray.TrayRecord{}, false, nil
}

func (s *apiSource) findLocalItem(id string) (localtray.ItemWithTray, bool, error) {
	return localtray.NewStore(s.configDir()).FindItem(id)
}

func fromLocal(item localtray.Item, rec localtray.TrayRecord) api.Item {
	return api.Item{
		ID:        item.ID,
		Scope:     "local",
		TrayRef:   bar.LocalTrayRef(rec),
		Title:     item.Title,
		Status:    item.Status,
		CreatedAt: item.CreatedAt,
		Body:      item.Body,
		Actions:   api.ActionsFor("local", item.Status),
	}
}

func fromRemote(it domain.Item, trayRef string) api.Item {
	return api.Item{
		ID:          it.ID,
		Scope:       "remote",
		TrayRef:     trayRef,
		Title:       it.Title,
		Status:      it.Status,
		CreatedAt:   it.CreatedAt,
		SnoozeUntil: it.SnoozeUntil,
		Body:        it.Body,
		Actions:     api.ActionsFor("remote", it.Status),
	}
}

// pendingOwnedQuery matches the menu bar: pending items on owned trays.
// An empty TrayIDIn is not a filter, so zero owned trays use the sentinel id.
func pendingOwnedQuery(owned []domain.Tray) domain.ListItemsQuery {
	q := domain.ListItemsQuery{Status: "pending", OrderCreated: "desc"}
	ids := ownedTrayIDList(owned)
	if len(ids) == 0 {
		q.TrayID = noOwnedTraysTrayFilter
		return q
	}
	q.TrayIDIn = ids
	return q
}

func ownedTrayIDList(trays []domain.Tray) []string {
	ids := make([]string, 0, len(trays))
	for i := range trays {
		if id := strings.TrimSpace(trays[i].ID); id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func resolveOwnedTray(owned []domain.Tray, ref string) (domain.Tray, error) {
	id, err := trayref.TrayIDFromRef(ref, nil, owned)
	if err != nil {
		return domain.Tray{}, api.ErrNotFound
	}
	t, ok := trayref.TrayByID(owned, id)
	if !ok {
		return domain.Tray{}, api.ErrNotFound
	}
	return t, nil
}

func remoteOpenStatus(status string) bool {
	switch status {
	case "pending", "accepted", "snoozed":
		return true
	default:
		return false
	}
}

func parseSnoozeUntil(until string) (time.Time, error) {
	until = strings.TrimSpace(until)
	t, err := time.Parse(time.RFC3339, until)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, until)
	}
	if err != nil || !t.After(time.Now()) {
		return time.Time{}, api.Invalid("until must be a future RFC3339 time")
	}
	return t, nil
}

func strPtr(s string) *string { return &s }

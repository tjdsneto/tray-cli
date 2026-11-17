package commands

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/api"
	"github.com/tjdsneto/tray-cli/internal/bar"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/localtray"
	"github.com/tjdsneto/tray-cli/internal/output"
	"github.com/tjdsneto/tray-cli/internal/remotesfile"
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
		return local, nil
	}
	return append(local, remote...), nil
}

func (s *apiSource) Items(ctx context.Context, trayRef string) ([]api.Item, error) {
	rec, ok, err := s.findLocal(trayRef)
	if err != nil {
		return nil, err
	}
	if ok {
		rows, err := localtray.NewStore(s.configDir()).ListAllItems([]string{rec.ID})
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
	return withSourceUser(fromRemote(it, ref), it, sess, profileDisplayMap(ctx, sess, svcs, sourceUserIDsFromItems([]domain.Item{it}))), nil
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
	tray, err := s.findRemoteTray(ctx, svcs, sess, req.Tray)
	if err != nil {
		if errors.Is(err, api.ErrNotFound) {
			return api.Item{}, false, nil
		}
		return api.Item{}, true, err
	}
	created, err := svcs.Items.Add(ctx, sess, tray.ID, strings.TrimSpace(req.Title), nil, nil, req.Body)
	if err != nil || created == nil {
		return api.Item{}, true, api.ErrRemote
	}
	ref := strings.TrimSpace(tray.Name)
	if ref == "" {
		ref = tray.ID
	}
	return withSourceUser(fromRemote(*created, ref), *created, sess, nil), true, nil
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
		name, detail := trayPresentation(sum.Record)
		out = append(out, api.Tray{
			Scope:     "local",
			Kind:      sum.Record.Kind,
			Ref:       bar.LocalTrayRef(sum.Record),
			Name:      name,
			Detail:    detail,
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
	joined, err := svcs.Trays.ListJoined(ctx, sess)
	if err != nil {
		return nil, api.ErrRemote
	}
	counts := map[string]int{}
	ids := append(ownedTrayIDList(owned), ownedTrayIDList(joined)...)
	if len(ids) > 0 {
		items, err := svcs.Items.List(ctx, sess, domain.ListItemsQuery{TrayIDIn: ids})
		if err == nil {
			for _, it := range items {
				if remoteOpenStatus(it.Status) {
					counts[it.TrayID]++
				}
			}
		}
	}
	out := make([]api.Tray, 0, len(owned)+len(joined))
	for _, t := range owned {
		out = append(out, remoteTray(t, "remote", counts[t.ID]))
	}
	aliases := aliasByTrayID(s.configDir())
	labels := profileDisplayMap(ctx, sess, svcs, ownerIDs(joined))
	for _, t := range joined {
		tray := remoteTray(t, "joined", counts[t.ID])
		tray.Detail = aliases[t.ID]
		tray.Owner = strings.TrimSpace(labels[strings.TrimSpace(t.OwnerID)])
		out = append(out, tray)
	}
	return out, nil
}

func ownerIDs(trays []domain.Tray) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, tray := range trays {
		id := strings.TrimSpace(tray.OwnerID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func aliasByTrayID(configDir string) map[string]string {
	f, err := remotesfile.Load(configDir)
	if err != nil {
		return map[string]string{}
	}
	out := map[string]string{}
	for alias, id := range f.Aliases {
		id = strings.TrimSpace(id)
		if id == "" || out[id] != "" {
			continue
		}
		out[id] = alias
	}
	return out
}

func remoteTray(t domain.Tray, kind string, open int) api.Tray {
	name := strings.TrimSpace(t.Name)
	if name == "" {
		name = t.ID
	}
	return api.Tray{
		Scope:     "remote",
		Kind:      kind,
		Ref:       t.ID,
		Name:      name,
		OpenCount: open,
	}
}

func (s *apiSource) remoteItems(ctx context.Context, trayRef string) ([]api.Item, error) {
	svcs, sess, err := s.auth()
	if err != nil {
		return nil, api.ErrSignedOut
	}
	tray, err := s.findRemoteTray(ctx, svcs, sess, trayRef)
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
	labels := profileDisplayMap(ctx, sess, svcs, sourceUserIDsFromItems(items))
	out := make([]api.Item, 0, len(items))
	for _, it := range items {
		out = append(out, withSourceUser(fromRemote(it, name), it, sess, labels))
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

func trayPresentation(rec localtray.TrayRecord) (name, detail string) {
	switch rec.Kind {
	case localtray.KindDir:
		p := strings.TrimSpace(rec.Path)
		base := filepath.Base(p)
		if base == "" || base == "." || base == string(filepath.Separator) {
			base = localtray.DisplayLabel(rec.ID)
		}
		return base, p
	case localtray.KindBranch:
		root := strings.TrimSpace(rec.RepoRoot)
		base := filepath.Base(root)
		if base == "" || base == "." || base == string(filepath.Separator) {
			base = localtray.DisplayLabel(rec.ID)
		}
		return base, strings.TrimSpace(rec.Branch)
	default:
		name = strings.TrimSpace(rec.Name)
		if name == "" {
			name = localtray.DisplayLabel(rec.ID)
		}
		return name, ""
	}
}

func fromLocal(item localtray.Item, rec localtray.TrayRecord) api.Item {
	ref := bar.LocalTrayRef(rec)
	return api.Item{
		ID:        item.ID,
		Scope:     "local",
		TrayRef:   ref,
		Title:     item.Title,
		Status:    item.Status,
		CreatedAt: item.CreatedAt,
		DoneAt:    item.DoneAt,
		Body:      item.Body,
		Clipboard: bar.ClipboardLine(bar.Row{TrayRef: ref, ItemID: item.ID}),
		Actions:   api.ActionsFor("local", item.Status),
	}
}

func fromRemote(it domain.Item, trayRef string) api.Item {
	return api.Item{
		ID:                it.ID,
		Scope:             "remote",
		TrayRef:           trayRef,
		Title:             it.Title,
		Status:            it.Status,
		CreatedAt:         it.CreatedAt,
		UpdatedAt:         timePtr(it.UpdatedAt),
		DueDate:           it.DueDate,
		SnoozeUntil:       it.SnoozeUntil,
		DeclineReason:     it.DeclineReason,
		CompletionMessage: it.CompletionMessage,
		AcceptedAt:        it.AcceptedAt,
		DeclinedAt:        it.DeclinedAt,
		CompletedAt:       it.CompletedAt,
		ArchivedAt:        it.ArchivedAt,
		Body:              it.Body,
		Clipboard:         bar.ClipboardLine(bar.Row{TrayRef: trayRef, ItemID: it.ID}),
		AgentSessionID:    it.AgentSessionID,
		Actions:           api.ActionsFor("remote", it.Status),
	}
}

func withSourceUser(item api.Item, raw domain.Item, sess domain.Session, labels map[string]string) api.Item {
	label := output.FormatSourceUser(raw.SourceUserID, sess.UserID, labels)
	if label != "" && label != "—" {
		item.SourceUser = label
	}
	return item
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
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

func (s *apiSource) findRemoteTray(ctx context.Context, svcs domain.Services, sess domain.Session, ref string) (domain.Tray, error) {
	owned, err := svcs.Trays.ListOwned(ctx, sess)
	if err != nil {
		return domain.Tray{}, api.ErrRemote
	}
	if tray, err := resolveOwnedTray(owned, ref); err == nil {
		return tray, nil
	}
	joined, err := svcs.Trays.ListJoined(ctx, sess)
	if err != nil {
		return domain.Tray{}, api.ErrRemote
	}
	return resolveOwnedTray(joined, ref)
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

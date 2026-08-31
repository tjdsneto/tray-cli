package localtray

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGlobalID_and_ParseRef(t *testing.T) {
	id, err := GlobalID("Inbox")
	require.NoError(t, err)
	require.Equal(t, "global:inbox", id)

	got, err := ParseRef("inbox", "/tmp", GitInfo{})
	require.NoError(t, err)
	require.Equal(t, id, got)
}

func TestDirID_and_AncestorDirs(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "proj", "pkg")
	require.NoError(t, os.MkdirAll(sub, 0o755))

	id, err := DirID(sub)
	require.NoError(t, err)
	require.Contains(t, id, "dir:")
	require.Contains(t, id, "proj")

	ancestors := AncestorDirs(sub)
	require.GreaterOrEqual(t, len(ancestors), 2)
	require.Equal(t, sub, ancestors[0])
}

func TestBranchID(t *testing.T) {
	root := t.TempDir()
	id, err := BranchID(root, "feat/auth")
	require.NoError(t, err)
	require.Equal(t, "branch:"+filepath.Clean(root)+":feat/auth", id)
}

func TestContextTrayIDs_order(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "repo")
	require.NoError(t, os.MkdirAll(sub, 0o755))

	dirID, err := DirID(sub)
	require.NoError(t, err)
	branchID, err := BranchID(sub, "dev")
	require.NoError(t, err)
	globalID, err := GlobalID("inbox")
	require.NoError(t, err)

	idx := Index{Trays: []TrayRecord{
		{ID: globalID, Kind: KindGlobal, Name: "inbox", ItemsFile: "items/a.jsonl"},
		{ID: dirID, Kind: KindDir, Path: sub, ItemsFile: "items/b.jsonl"},
		{ID: branchID, Kind: KindBranch, RepoRoot: sub, Branch: "dev", ItemsFile: "items/c.jsonl"},
	}}

	git := GitInfo{InRepo: true, RepoRoot: sub, Branch: "dev", IsDefaultBranch: false}
	ids := ContextTrayIDs(idx, sub, git)
	require.Equal(t, []string{branchID, dirID, globalID}, ids)
}

func TestDefaultAddTarget_branchVsDir(t *testing.T) {
	root := t.TempDir()
	git := GitInfo{InRepo: true, RepoRoot: root, Branch: "feat", IsDefaultBranch: false}
	target, err := DefaultAddTarget(root, git)
	require.NoError(t, err)
	require.Equal(t, KindBranch, target.Kind)

	git = GitInfo{InRepo: true, RepoRoot: root, Branch: "main", IsDefaultBranch: true}
	target, err = DefaultAddTarget(root, git)
	require.NoError(t, err)
	require.Equal(t, KindDir, target.Kind)
	require.Equal(t, root, target.Path)
}

func TestStore_add_list_prune(t *testing.T) {
	cfg := t.TempDir()
	store := NewStore(cfg)
	root := cfg

	item, rec, err := store.AddItem(AddTarget{Kind: KindGlobal, Name: "inbox"}, "hello", true)
	require.NoError(t, err)
	require.NotEmpty(t, item.ID)
	require.Equal(t, "global:inbox", rec.ID)

	rows, err := store.ListItemsForTrays([]string{rec.ID})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "hello", rows[0].Item.Title)

	pruned, err := store.PruneEmpty(false)
	require.NoError(t, err)
	require.Empty(t, pruned)

	_, err = store.CompleteItem(item.ID)
	require.NoError(t, err)
	rows, err = store.ListItemsForTrays([]string{rec.ID})
	require.NoError(t, err)
	require.Empty(t, rows)

	dir := filepath.Join(root, "empty-dir")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	emptyTarget := AddTarget{Kind: KindDir, Path: dir}
	recEmpty, err := store.EnsureFromTarget(emptyTarget)
	require.NoError(t, err)
	require.NotEmpty(t, recEmpty.ID)

	pruned, err = store.PruneEmpty(false)
	require.NoError(t, err)
	require.Contains(t, pruned[0], "empty-dir")
}

func TestStore_noCreate(t *testing.T) {
	store := NewStore(t.TempDir())
	_, _, err := store.AddItem(AddTarget{Kind: KindGlobal, Name: "x"}, "nope", false)
	require.Error(t, err)
}

func TestResolveItemPrefix(t *testing.T) {
	store := NewStore(t.TempDir())
	item, _, err := store.AddItem(AddTarget{Kind: KindGlobal, Name: "t"}, "one", true)
	require.NoError(t, err)

	id, err := store.ResolveItemPrefix(item.ID[:8])
	require.NoError(t, err)
	require.Equal(t, item.ID, id)
}

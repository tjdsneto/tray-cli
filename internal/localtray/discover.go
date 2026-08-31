package localtray

import (
	"path/filepath"
	"strings"
)

// AddTarget describes where a default add should land.
type AddTarget struct {
	Kind     string
	Name     string // global
	Path     string // dir
	RepoRoot string // branch
	Branch   string
}

// AncestorDirs returns cwd then each parent up to filesystem root.
func AncestorDirs(cwd string) []string {
	abs, err := filepath.Abs(strings.TrimSpace(cwd))
	if err != nil {
		return nil
	}
	abs = filepath.Clean(abs)
	var out []string
	for {
		out = append(out, abs)
		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}
	return out
}

// DefaultAddTarget picks branch tray on non-default branches, otherwise cwd directory tray.
func DefaultAddTarget(cwd string, git GitInfo) (AddTarget, error) {
	if git.InRepo && git.Branch != "" && !git.IsDefaultBranch {
		return AddTarget{Kind: KindBranch, RepoRoot: git.RepoRoot, Branch: git.Branch}, nil
	}
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return AddTarget{}, err
	}
	return AddTarget{Kind: KindDir, Path: filepath.Clean(abs)}, nil
}

// ContextTrayIDs returns tray ids for the default list view: branch (if indexed), ancestor dirs
// (deepest first, indexed only), then globals (indexed only). Order is most-specific first.
func ContextTrayIDs(idx Index, cwd string, git GitInfo) []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(id string) {
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		if _, ok := idx.Find(id); !ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}

	if git.InRepo && git.Branch != "" {
		if id, err := BranchID(git.RepoRoot, git.Branch); err == nil {
			add(id)
		}
	}
	for _, d := range AncestorDirs(cwd) {
		if id, err := DirID(d); err == nil {
			add(id)
		}
	}
	for _, rec := range idx.Globals() {
		add(rec.ID)
	}
	return out
}

// AllTrayIDs returns every tray id in index order (globals, dirs, branches — stable file order).
func AllTrayIDs(idx Index) []string {
	out := make([]string, 0, len(idx.Trays))
	for i := range idx.Trays {
		out = append(out, idx.Trays[i].ID)
	}
	return out
}

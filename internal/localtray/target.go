package localtray

import (
	"fmt"
	"os"
	"path/filepath"
)

// TargetFromFlags resolves an add target from explicit flags (at most one scope flag should be set).
func TargetFromFlags(here, branch, project bool, cwd string, git GitInfo) (AddTarget, error) {
	set := 0
	if here {
		set++
	}
	if branch {
		set++
	}
	if project {
		set++
	}
	if set > 1 {
		return AddTarget{}, fmt.Errorf("use only one of --here, --branch, or --project")
	}
	switch {
	case here:
		abs, err := filepath.Abs(cwd)
		if err != nil {
			return AddTarget{}, err
		}
		return AddTarget{Kind: KindDir, Path: abs}, nil
	case branch:
		if !git.InRepo || git.Branch == "" {
			return AddTarget{}, fmt.Errorf("--branch requires a git repository with a named branch")
		}
		return AddTarget{Kind: KindBranch, RepoRoot: git.RepoRoot, Branch: git.Branch}, nil
	case project:
		if git.InRepo {
			return AddTarget{Kind: KindDir, Path: git.RepoRoot}, nil
		}
		abs, err := filepath.Abs(cwd)
		if err != nil {
			return AddTarget{}, err
		}
		return AddTarget{Kind: KindDir, Path: abs}, nil
	default:
		return DefaultAddTarget(cwd, git)
	}
}

// TargetFromRef resolves a named tray reference for local add/list.
func TargetFromRef(ref string, cwd string, git GitInfo) (AddTarget, error) {
	id, err := ParseRef(ref, cwd, git)
	if err != nil {
		return AddTarget{}, err
	}
	switch {
	case hasPrefix(id, KindGlobal+":"):
		name := id[len(KindGlobal)+1:]
		return AddTarget{Kind: KindGlobal, Name: name}, nil
	case hasPrefix(id, KindDir+":"):
		path := id[len(KindDir)+1:]
		return AddTarget{Kind: KindDir, Path: path}, nil
	case hasPrefix(id, KindBranch+":"):
		rest := id[len(KindBranch)+1:]
		i := lastColon(rest)
		if i < 0 {
			return AddTarget{}, fmt.Errorf("invalid branch tray id %q", id)
		}
		return AddTarget{
			Kind:     KindBranch,
			RepoRoot: rest[:i],
			Branch:   rest[i+1:],
		}, nil
	default:
		return AddTarget{}, fmt.Errorf("unknown tray id %q", id)
	}
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func lastColon(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == ':' {
			return i
		}
	}
	return -1
}

// WorkingCwd returns os.Getwd() or "." on error.
func WorkingCwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

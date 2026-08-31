package localtray

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	KindGlobal = "global"
	KindDir    = "dir"
	KindBranch = "branch"
)

// GlobalID returns the canonical id for a named global tray.
func GlobalID(name string) (string, error) {
	n := normalizeGlobalName(name)
	if n == "" {
		return "", fmt.Errorf("global tray name cannot be empty")
	}
	return KindGlobal + ":" + n, nil
}

// DirID returns the canonical id for a directory tray (absolute, cleaned path).
func DirID(absPath string) (string, error) {
	p, err := filepath.Abs(strings.TrimSpace(absPath))
	if err != nil {
		return "", err
	}
	p = filepath.Clean(p)
	if p == "" || p == "." {
		return "", fmt.Errorf("directory path cannot be empty")
	}
	return KindDir + ":" + p, nil
}

// BranchID returns the canonical id for a branch tray.
func BranchID(repoRoot, branch string) (string, error) {
	root, err := filepath.Abs(strings.TrimSpace(repoRoot))
	if err != nil {
		return "", err
	}
	root = filepath.Clean(root)
	b := strings.TrimSpace(branch)
	if root == "" || root == "." {
		return "", fmt.Errorf("repository root cannot be empty")
	}
	if b == "" {
		return "", fmt.Errorf("branch name cannot be empty")
	}
	return KindBranch + ":" + root + ":" + b, nil
}

func normalizeGlobalName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// ParseRef resolves a user-provided tray reference to a canonical id when unambiguous.
// Supports global names, dir:/path, branch:name (within repoRoot when inRepo), and full ids.
func ParseRef(ref string, cwd string, git GitInfo) (string, error) {
	r := strings.TrimSpace(ref)
	if r == "" {
		return "", fmt.Errorf("empty tray reference")
	}
	lower := strings.ToLower(r)
	if strings.HasPrefix(lower, KindGlobal+":") || strings.HasPrefix(lower, KindDir+":") || strings.HasPrefix(lower, KindBranch+":") {
		return r, nil
	}
	// Bare name → global tray.
	return GlobalID(r)
}

// DisplayLabel formats a tray id for human output.
func DisplayLabel(id string) string {
	id = strings.TrimSpace(id)
	switch {
	case strings.HasPrefix(id, KindGlobal+":"):
		return "[global] " + strings.TrimPrefix(id, KindGlobal+":")
	case strings.HasPrefix(id, KindDir+":"):
		p := strings.TrimPrefix(id, KindDir+":")
		return "[dir] " + filepath.Base(p)
	case strings.HasPrefix(id, KindBranch+":"):
		rest := strings.TrimPrefix(id, KindBranch+":")
		i := strings.LastIndex(rest, ":")
		if i < 0 {
			return "[branch] " + rest
		}
		repo := rest[:i]
		branch := rest[i+1:]
		return fmt.Sprintf("[branch:%s] %s", branch, filepath.Base(repo))
	default:
		return id
	}
}

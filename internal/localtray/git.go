package localtray

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// GitInfo describes the git context for the current working directory.
type GitInfo struct {
	InRepo          bool
	RepoRoot        string
	Branch          string
	IsDefaultBranch bool
}

// DetectGit runs git to resolve repository root and current branch from cwd.
func DetectGit(cwd string) GitInfo {
	cwd = strings.TrimSpace(cwd)
	if cwd == "" {
		return GitInfo{}
	}
	root, ok := gitOutput(cwd, "rev-parse", "--show-toplevel")
	if !ok {
		return GitInfo{}
	}
	root = filepath.Clean(root)
	branch, ok := gitOutput(cwd, "rev-parse", "--abbrev-ref", "HEAD")
	if !ok || branch == "" || branch == "HEAD" {
		return GitInfo{InRepo: true, RepoRoot: root}
	}
	return GitInfo{
		InRepo:          true,
		RepoRoot:        root,
		Branch:          branch,
		IsDefaultBranch: isDefaultBranch(branch),
	}
}

func gitOutput(dir string, args ...string) (string, bool) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}

func isDefaultBranch(branch string) bool {
	switch strings.ToLower(strings.TrimSpace(branch)) {
	case "main", "master":
		return true
	default:
		return false
	}
}

package config

import "strings"

// Version is the release identifier (e.g. v1.2.3). Set at link time for release builds; dev builds use "dev".
var Version = "dev"

// GitCommit is the short git revision, set at link time for release builds.
var GitCommit = ""

// VersionLine returns a single line for --version (and similar).
func VersionLine() string {
	v := strings.TrimSpace(Version)
	if v == "" {
		v = "dev"
	}
	c := strings.TrimSpace(GitCommit)
	if c == "" {
		return v
	}
	return v + " (" + c + ")"
}

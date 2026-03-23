package commands

import "github.com/tjdsneto/tray-cli/internal/domain"

// Deps is wiring from package cli (auth, paths) so this package stays import-cycle free.
type Deps struct {
	RequireAuth   func() (domain.Services, domain.Session, error)
	ConfigDir     func() string
	RemoteAliases func() map[string]string
}

var cmdDeps Deps

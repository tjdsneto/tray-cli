package domain

import "errors"

// ErrNotImplemented is returned by service methods not wired up yet.
var ErrNotImplemented = errors.New("domain: not implemented yet")

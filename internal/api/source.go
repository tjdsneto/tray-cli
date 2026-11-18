package api

import (
	"context"
	"errors"

	"github.com/tjdsneto/tray-cli/internal/bar"
)

type Ctx = context.Context

var ErrNotFound = errors.New("not found")
var ErrRemote = errors.New("remote unavailable")
var ErrSignedOut = errors.New("signed out")

type ActBody struct {
	Reason string
	Until  string
}

type Source interface {
	SignedIn(ctx Ctx) (bool, error)
	// LocalAttention is the disk-only half of Attention. It does not call remote.
	LocalAttention(ctx Ctx) ([]bar.Item, error)
	Attention(ctx Ctx) (local []bar.Item, remote []bar.Item, remoteErr error)
	Trays(ctx Ctx) ([]Tray, error)
	Items(ctx Ctx, trayRef string) ([]Item, error)
	Item(ctx Ctx, id string) (Item, error)
	Add(ctx Ctx, req AddRequest) (Item, error)
	Act(ctx Ctx, id, action string, body ActBody) error
}

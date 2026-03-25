//go:build !unix

package commands

import "context"

func notifySIGHUP(_ context.Context, _ func()) {}

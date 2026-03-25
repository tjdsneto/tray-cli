package commands

import (
	"sync/atomic"

	"github.com/tjdsneto/tray-cli/internal/cli/listenhook"
)

// hookRuntime is an immutable snapshot of merged hook rules (file + --exec).
type hookRuntime struct {
	cfg           *listenhook.Config
	ruleTray      []string
	hookPathLabel string
}

func (h *hookRuntime) snapshot() (*listenhook.Config, []string, string) {
	if h == nil {
		return nil, nil, ""
	}
	return h.cfg, h.ruleTray, h.hookPathLabel
}

func snapHooks(ap *atomic.Pointer[hookRuntime]) (*listenhook.Config, []string, string) {
	if ap == nil {
		return nil, nil, ""
	}
	h := ap.Load()
	if h == nil {
		return nil, nil, ""
	}
	return h.snapshot()
}

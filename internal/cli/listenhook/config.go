package listenhook

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Well-known event names for hook rules.
const (
	EventItemPending   = "item.pending"
	EventItemCompleted = "item.completed"
	EventItemAccepted  = "item.accepted"
	EventItemDeclined  = "item.declined"
)

// Scope values for hook rules.
const (
	ScopeInboxOwned = "inbox_owned"
	ScopeOutbox     = "outbox"
)

// Config is the Claude-style hooks file (JSON) for tray listen.
type Config struct {
	Hooks []Rule `json:"hooks"`
}

// Rule is one hook: when event fires and filters match, run command with listenhook env vars (see Env* in env.go).
type Rule struct {
	Event      string   `json:"event"`
	Scope      string   `json:"scope,omitempty"`
	FromOthers *bool    `json:"from_others,omitempty"`
	Tray       string   `json:"tray,omitempty"`
	Command    []string `json:"command"`
}

// Load reads and validates a hooks JSON file.
func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("listenhook: parse %s: %w", path, err)
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

// Validate checks rules; call after json.Unmarshal.
func (c *Config) Validate() error {
	for i := range c.Hooks {
		if err := c.Hooks[i].Validate(); err != nil {
			return fmt.Errorf("hooks[%d]: %w", i, err)
		}
	}
	return nil
}

// Validate checks one rule.
func (r *Rule) Validate() error {
	ev := strings.TrimSpace(r.Event)
	switch ev {
	case EventItemPending, EventItemCompleted, EventItemAccepted, EventItemDeclined:
	default:
		return fmt.Errorf("unsupported event %q", r.Event)
	}
	if len(r.Command) == 0 {
		return fmt.Errorf("missing command")
	}
	if strings.TrimSpace(r.Command[0]) == "" {
		return fmt.Errorf("empty command[0]")
	}
	switch ev {
	case EventItemPending:
		sc := pendingScope(r)
		switch sc {
		case ScopeInboxOwned, "":
		default:
			return fmt.Errorf("item.pending: unsupported scope %q", r.Scope)
		}
	case EventItemCompleted, EventItemAccepted, EventItemDeclined:
		sc := outboxScope(r)
		switch sc {
		case ScopeOutbox, "":
		default:
			return fmt.Errorf("%s: unsupported scope %q (use outbox or omit)", ev, r.Scope)
		}
	}
	return nil
}

// WantsPendingPoll returns true if any rule needs the pending items query.
func (c *Config) WantsPendingPoll() bool {
	for _, r := range c.Hooks {
		if strings.TrimSpace(r.Event) == EventItemPending {
			return true
		}
	}
	return false
}

// WantsOutboxPoll returns true if any rule needs ListOutbox.
func (c *Config) WantsOutboxPoll() bool {
	for _, r := range c.Hooks {
		switch strings.TrimSpace(r.Event) {
		case EventItemCompleted, EventItemAccepted, EventItemDeclined:
			return true
		}
	}
	return false
}

func pendingScope(r *Rule) string {
	s := strings.TrimSpace(r.Scope)
	if s == "" {
		return ScopeInboxOwned
	}
	return s
}

func outboxScope(r *Rule) string {
	s := strings.TrimSpace(r.Scope)
	if s == "" {
		return ScopeOutbox
	}
	return s
}

// FromOthersDefault returns whether to exclude self-sourced items for pending inbox rules.
func (r *Rule) FromOthersDefault() bool {
	if r.FromOthers != nil {
		return *r.FromOthers
	}
	return true
}

package listenhook

import (
	"strings"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

// PendingSeen tracks item IDs already observed in the pending poll (in-memory only).
type PendingSeen struct {
	seen map[string]struct{}
}

// NewPendingSeen returns an empty tracker.
func NewPendingSeen() *PendingSeen {
	return &PendingSeen{seen: make(map[string]struct{})}
}

// Seed marks ids as seen without returning them as new (initial snapshot).
func (p *PendingSeen) Seed(items []domain.Item) {
	if p.seen == nil {
		p.seen = make(map[string]struct{})
	}
	for _, it := range items {
		id := strings.TrimSpace(it.ID)
		if id == "" {
			continue
		}
		p.seen[id] = struct{}{}
	}
}

// NewPending returns items whose id was not in seen; each id is then marked seen.
func (p *PendingSeen) NewPending(items []domain.Item) []domain.Item {
	if p.seen == nil {
		p.seen = make(map[string]struct{})
	}
	var out []domain.Item
	for _, it := range items {
		id := strings.TrimSpace(it.ID)
		if id == "" {
			continue
		}
		if _, ok := p.seen[id]; ok {
			continue
		}
		p.seen[id] = struct{}{}
		out = append(out, it)
	}
	return out
}

// OutboxState tracks last known status per outbox item id (in-memory only).
type OutboxState struct {
	last map[string]string
}

// NewOutboxState returns an empty tracker.
func NewOutboxState() *OutboxState {
	return &OutboxState{last: make(map[string]string)}
}

// Seed records current statuses without emitting transitions.
func (o *OutboxState) Seed(items []domain.Item) {
	if o.last == nil {
		o.last = make(map[string]string)
	}
	for _, it := range items {
		id := strings.TrimSpace(it.ID)
		if id == "" {
			continue
		}
		o.last[id] = strings.TrimSpace(it.Status)
	}
}

// OutboxTransitions lists outbox items that changed status into completed, accepted, or declined.
type OutboxTransitions struct {
	Completed []domain.Item
	Accepted  []domain.Item
	Declined  []domain.Item
}

// OutTransitions returns items that transitioned into completed, accepted, or declined (single pass; updates last).
func (o *OutboxState) OutTransitions(items []domain.Item) OutboxTransitions {
	if o.last == nil {
		o.last = make(map[string]string)
	}
	var t OutboxTransitions
	for _, it := range items {
		id := strings.TrimSpace(it.ID)
		if id == "" {
			continue
		}
		prev, had := o.last[id]
		cur := strings.TrimSpace(it.Status)
		o.last[id] = cur
		if !had {
			continue
		}
		if strings.EqualFold(prev, cur) {
			continue
		}
		if strings.EqualFold(cur, "completed") && !strings.EqualFold(prev, "completed") {
			t.Completed = append(t.Completed, it)
		}
		if strings.EqualFold(cur, "accepted") && !strings.EqualFold(prev, "accepted") {
			t.Accepted = append(t.Accepted, it)
		}
		if strings.EqualFold(cur, "declined") && !strings.EqualFold(prev, "declined") {
			t.Declined = append(t.Declined, it)
		}
	}
	return t
}

package listenhook

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()
	falseVal := false
	c := &Config{Hooks: []Rule{
		{Event: EventItemPending, Command: []string{"echo"}},
		{Event: EventItemCompleted, Command: []string{"/bin/true"}},
		{Event: EventItemAccepted, Command: []string{"/bin/true"}},
		{Event: EventItemDeclined, Command: []string{"/bin/true"}},
		{Event: EventItemPending, Scope: ScopeInboxOwned, FromOthers: &falseVal, Command: []string{"sh"}},
	}}
	require.NoError(t, c.Validate())

	bad := &Config{Hooks: []Rule{{Event: "nope", Command: []string{"x"}}}}
	require.Error(t, bad.Validate())
}

func TestPendingSeen_NewPending(t *testing.T) {
	t.Parallel()
	p := NewPendingSeen()
	items := []domain.Item{{ID: "1"}, {ID: "2"}}
	p.Seed(items)
	got := p.NewPending([]domain.Item{{ID: "1"}, {ID: "2"}, {ID: "3"}})
	require.Len(t, got, 1)
	require.Equal(t, "3", got[0].ID)
	require.Empty(t, p.NewPending([]domain.Item{{ID: "1"}, {ID: "2"}, {ID: "3"}}))
}

func TestOutboxState_OutTransitions(t *testing.T) {
	t.Parallel()
	o := NewOutboxState()
	now := time.Now()
	o.Seed([]domain.Item{
		{ID: "a", Status: "pending"},
		{ID: "b", Status: "completed"},
	})
	z := o.OutTransitions([]domain.Item{
		{ID: "a", Status: "pending"},
		{ID: "b", Status: "completed"},
	})
	require.Empty(t, z.Completed)
	require.Empty(t, z.Accepted)
	require.Empty(t, z.Declined)

	t1 := o.OutTransitions([]domain.Item{
		{ID: "a", Status: "completed", CompletedAt: &now},
	})
	require.Len(t, t1.Completed, 1)
	require.Equal(t, "a", t1.Completed[0].ID)

	o2 := NewOutboxState()
	o2.Seed([]domain.Item{{ID: "x", Status: "pending"}})
	t2 := o2.OutTransitions([]domain.Item{{ID: "x", Status: "accepted"}})
	require.Len(t, t2.Accepted, 1)
	require.Equal(t, "x", t2.Accepted[0].ID)

	o3 := NewOutboxState()
	o3.Seed([]domain.Item{{ID: "y", Status: "pending"}})
	reason := "too busy"
	t3 := o3.OutTransitions([]domain.Item{{ID: "y", Status: "declined", DeclineReason: &reason}})
	require.Len(t, t3.Declined, 1)
	require.Equal(t, "y", t3.Declined[0].ID)
}

func TestMatchPending(t *testing.T) {
	t.Parallel()
	owned := map[string]bool{"tr1": true}
	r := Rule{Event: EventItemPending, Scope: ScopeInboxOwned, Command: []string{"true"}}
	it := domain.Item{ID: "i", TrayID: "tr1", SourceUserID: "other", Status: "pending"}
	require.True(t, MatchPending(r, it, owned, "me", ""))

	self := domain.Item{ID: "i2", TrayID: "tr1", SourceUserID: "me", Status: "pending"}
	require.False(t, MatchPending(r, self, owned, "me", ""))

	falseOthers := false
	r2 := Rule{Event: EventItemPending, Scope: ScopeInboxOwned, FromOthers: &falseOthers, Command: []string{"true"}}
	require.True(t, MatchPending(r2, self, owned, "me", ""))
}

func TestHookEnv(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	it := domain.Item{
		ID: "id1", TrayID: "t1", Title: "hi", Status: "completed",
		SourceUserID: "su", CompletedAt: &now,
	}
	env := HookEnv(EventItemCompleted, domain.Session{UserID: "u1"}, it, "Pat Example")
	require.Contains(t, env, EnvHookEvent+"=item.completed")
	require.Contains(t, env, EnvSessionUserID+"=u1")
	require.Contains(t, env, EnvItemID+"=id1")
	require.Contains(t, env, EnvItemAddedByDisplayName+"=Pat Example")
	require.Contains(t, env, EnvItemCompletedAt+"="+now.UTC().Format(time.RFC3339Nano))
	require.Contains(t, env, EnvItemDeclineReason+"=")
}

func TestHookEnv_DeclineReason(t *testing.T) {
	t.Parallel()
	reason := "line1\nline2"
	it := domain.Item{ID: "i", Status: "declined", DeclineReason: &reason}
	env := HookEnv(EventItemDeclined, domain.Session{}, it, "")
	require.Contains(t, env, EnvItemDeclineReason+"=line1 line2")
}

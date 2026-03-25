package listenhook

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

// Environment variable names passed to hook commands (prefix TRAY_ avoids collisions).
const (
	EnvHookEvent = "TRAY_HOOK_EVENT" // e.g. item.pending, item.completed

	EnvSessionUserID = "TRAY_SESSION_USER_ID" // signed-in CLI user

	EnvItemID                 = "TRAY_ITEM_ID"
	EnvItemTrayID             = "TRAY_ITEM_TRAY_ID" // tray this line item belongs to
	EnvItemTitle              = "TRAY_ITEM_TITLE"
	EnvItemStatus             = "TRAY_ITEM_STATUS"
	EnvItemAddedByUserID      = "TRAY_ITEM_ADDED_BY_USER_ID"
	EnvItemAddedByDisplayName = "TRAY_ITEM_ADDED_BY_DISPLAY_NAME" // profile name or email when available

	EnvItemCompletedAt = "TRAY_ITEM_COMPLETED_AT"
	EnvItemAcceptedAt  = "TRAY_ITEM_ACCEPTED_AT"
	EnvItemDeclinedAt  = "TRAY_ITEM_DECLINED_AT"
	// DeclineReason is set when the owner declined (owner-provided text; newlines replaced with spaces for env safety).
	EnvItemDeclineReason = "TRAY_ITEM_DECLINE_REASON"
)

// HookEnv returns environment key=value pairs to append to the process environment.
// sourceDisplayName is the resolved label for the user who added the item (e.g. profile name or email), or empty if unknown.
func HookEnv(event string, sess domain.Session, it domain.Item, sourceDisplayName string) []string {
	uid := strings.TrimSpace(sess.UserID)
	var pairs []string
	add := func(k, v string) {
		pairs = append(pairs, k+"="+v)
	}
	add(EnvHookEvent, event)
	add(EnvSessionUserID, uid)
	add(EnvItemID, strings.TrimSpace(it.ID))
	add(EnvItemTrayID, strings.TrimSpace(it.TrayID))
	add(EnvItemTitle, it.Title)
	add(EnvItemStatus, strings.TrimSpace(it.Status))
	add(EnvItemAddedByUserID, strings.TrimSpace(it.SourceUserID))
	add(EnvItemAddedByDisplayName, strings.TrimSpace(sourceDisplayName))
	if it.CompletedAt != nil {
		add(EnvItemCompletedAt, it.CompletedAt.UTC().Format(time.RFC3339Nano))
	}
	if it.AcceptedAt != nil {
		add(EnvItemAcceptedAt, it.AcceptedAt.UTC().Format(time.RFC3339Nano))
	}
	if it.DeclinedAt != nil {
		add(EnvItemDeclinedAt, it.DeclinedAt.UTC().Format(time.RFC3339Nano))
	}
	add(EnvItemDeclineReason, declineReasonForEnv(it.DeclineReason))
	return pairs
}

func declineReasonForEnv(v *string) string {
	if v == nil {
		return ""
	}
	s := strings.TrimSpace(*v)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// MergeEnv appends hook pairs to base (typically os.Environ()).
func MergeEnv(base []string, hook []string) []string {
	out := make([]string, 0, len(base)+len(hook))
	out = append(out, base...)
	out = append(out, hook...)
	return out
}

// RunHook executes rule.Command with HookEnv merged into the environment.
func RunHook(rule Rule, event string, sess domain.Session, it domain.Item, sourceDisplayName string) error {
	cmd := rule.Command
	if len(cmd) == 0 {
		return fmt.Errorf("listenhook: empty command")
	}
	env := MergeEnv(os.Environ(), HookEnv(event, sess, it, sourceDisplayName))
	return runCmd(cmd[0], cmd[1:], env)
}

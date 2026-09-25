package agentsession

import (
	"fmt"
	"os"
	"strings"
)

// EnvAgentSessionID is the live agent session identity (list “mine”, add source default).
const EnvAgentSessionID = "TRAY_AGENT_SESSION_ID"

// ListAgentSessionIDFromEnv is the cobra NoOptDefVal for bare `tray list --agent-session-id`
// (“mine” via $TRAY_AGENT_SESSION_ID). Must be non-empty: pflag rejects NoOptDefVal="".
// With NoOptDefVal set, `--agent-session-id <id>` leaves <id> as a positional — use
// CoalesceListAgentSessionFlag before ResolveListAgentSessionID.
const ListAgentSessionIDFromEnv = "@env"

// FlagOrEnv returns trimmed flag if non-empty, else trimmed os.Getenv(envKey), else "".
func FlagOrEnv(flagVal, envKey string) string {
	if v := strings.TrimSpace(flagVal); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv(envKey))
}

// CoalesceListAgentSessionFlag restores space-separated `--agent-session-id <id>` when
// NoOptDefVal left the id as a positional argument. Returns the effective flag value and
// any remaining positional (empty when the positional was consumed as the session id).
func CoalesceListAgentSessionFlag(flagVal, positional string) (effectiveFlag, remainingPositional string) {
	flagVal = strings.TrimSpace(flagVal)
	positional = strings.TrimSpace(positional)
	if flagVal == ListAgentSessionIDFromEnv && positional != "" {
		return positional, ""
	}
	return flagVal, positional
}

// ResolveListAgentSessionID resolves the session id for `tray list --agent-session-id`.
// When flagChanged is false, returns active=false (normal list). When true, uses trimmed
// flagVal if non-empty and not ListAgentSessionIDFromEnv, else envVal; both empty → error.
func ResolveListAgentSessionID(flagChanged bool, flagVal, envVal string) (id string, active bool, err error) {
	if !flagChanged {
		return "", false, nil
	}
	id = strings.TrimSpace(flagVal)
	if id == "" || id == ListAgentSessionIDFromEnv {
		id = strings.TrimSpace(envVal)
	}
	if id == "" {
		return "", true, fmt.Errorf("pass --agent-session-id <id> or set %s", EnvAgentSessionID)
	}
	return id, true, nil
}

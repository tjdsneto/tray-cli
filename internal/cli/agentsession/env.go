package agentsession

import (
	"fmt"
	"os"
	"strings"
)

// EnvAgentSessionID is the live agent session identity (list “mine”, add source default).
const EnvAgentSessionID = "TRAY_AGENT_SESSION_ID"

// FlagOrEnv returns trimmed flag if non-empty, else trimmed os.Getenv(envKey), else "".
func FlagOrEnv(flagVal, envKey string) string {
	if v := strings.TrimSpace(flagVal); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv(envKey))
}

// ResolveListAgentSessionID resolves the session id for `tray list --agent-session-id`.
// When flagChanged is false, returns active=false (normal list). When true, uses trimmed
// flagVal if non-empty, else envVal; both empty → error.
func ResolveListAgentSessionID(flagChanged bool, flagVal, envVal string) (id string, active bool, err error) {
	if !flagChanged {
		return "", false, nil
	}
	id = strings.TrimSpace(flagVal)
	if id == "" {
		id = strings.TrimSpace(envVal)
	}
	if id == "" {
		return "", true, fmt.Errorf("pass --agent-session-id <id> or set %s", EnvAgentSessionID)
	}
	return id, true, nil
}

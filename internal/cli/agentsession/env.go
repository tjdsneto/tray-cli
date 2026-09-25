package agentsession

import (
	"os"
	"strings"
)

// FlagOrEnv returns trimmed flag if non-empty, else trimmed os.Getenv(envKey), else "".
func FlagOrEnv(flagVal, envKey string) string {
	if v := strings.TrimSpace(flagVal); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv(envKey))
}

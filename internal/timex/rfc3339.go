package timex

import (
	"fmt"
	"time"
)

// ParseRFC3339OrNano parses a non-empty timestamp string using RFC3339Nano, then RFC3339.
// It is suitable for JSON API fields (e.g. PostgREST / OpenAPI) where either layout may appear.
func ParseRFC3339OrNano(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("timex: empty timestamp")
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
	}
	return t, err
}

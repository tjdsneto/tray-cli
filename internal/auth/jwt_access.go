package auth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

// accessTokenExpiry returns the JWT exp claim (UTC) if present and parseable.
func accessTokenExpiry(accessToken string) (time.Time, bool) {
	parts := strings.Split(strings.TrimSpace(accessToken), ".")
	if len(parts) != 3 {
		return time.Time{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		// try padded URL encoding
		s := parts[1]
		if m := len(s) % 4; m != 0 {
			s += strings.Repeat("=", 4-m)
		}
		payload, err = base64.URLEncoding.DecodeString(s)
		if err != nil {
			return time.Time{}, false
		}
	}
	var p struct {
		Exp *float64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &p); err != nil || p.Exp == nil {
		return time.Time{}, false
	}
	return time.Unix(int64(*p.Exp), 0).UTC(), true
}

// accessTokenNeedsRefresh returns true if the token is missing, has no exp, or expires
// within leeway (refresh before the access JWT is rejected by the API).
func accessTokenNeedsRefresh(accessToken string, leeway time.Duration) bool {
	if strings.TrimSpace(accessToken) == "" {
		return true
	}
	exp, ok := accessTokenExpiry(accessToken)
	if !ok {
		return true
	}
	return time.Until(exp) < leeway
}

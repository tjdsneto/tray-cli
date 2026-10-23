package domain

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// MaxBodyRunes is the hard cap for item body text (Unicode runes).
const MaxBodyRunes = 100_000

// NormalizeBody trims space; empty → nil; over MaxBodyRunes → error.
func NormalizeBody(s string) (*string, error) {
	t := strings.TrimSpace(s)
	if t == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(t) > MaxBodyRunes {
		return nil, fmt.Errorf("body exceeds %d characters", MaxBodyRunes)
	}
	return &t, nil
}

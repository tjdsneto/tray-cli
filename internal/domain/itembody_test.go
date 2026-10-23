package domain

import (
	"strings"
	"testing"
)

func TestNormalizeBody(t *testing.T) {
	t.Parallel()

	if b, err := NormalizeBody("  "); err != nil || b != nil {
		t.Fatalf("empty: got (%v, %v)", b, err)
	}

	got, err := NormalizeBody("  hi  ")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || *got != "hi" {
		t.Fatalf("trim: got %v", got)
	}

	ok := strings.Repeat("a", MaxBodyRunes)
	if _, err := NormalizeBody(ok); err != nil {
		t.Fatal(err)
	}
	if _, err := NormalizeBody(ok + "x"); err == nil {
		t.Fatal("expected over-cap error")
	}
}

package twitch

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLowercaseAll(t *testing.T) {
	input := []string{
		"lower",
		"lower-with-symbols",
		"Mixed",
		"Mixed-with-Symbols",
		"UPPER",
		"UPPER-WITH-SYMBOLS",
	}
	want := []string{
		"lower",
		"lower-with-symbols",
		"mixed",
		"mixed-with-symbols",
		"upper",
		"upper-with-symbols",
	}

	got := lowercaseAll(input)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("lowercaseAll() diff (-want +got):\n%s", diff)
	}
}

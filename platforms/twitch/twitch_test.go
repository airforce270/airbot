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

func TestBypassSameMessageDetection(t *testing.T) {
	tests := []struct {
		desc  string
		input string
		want  string
	}{
		{
			desc:  "command with no later space",
			input: "/ban xqc",
			want:  "/ban xqc" + messageSpaceSuffix,
		},
		{
			desc:  "command with a later space",
			input: "/timeout xqc 100000",
			want:  "/timeout  xqc 100000",
		},
		{
			desc:  "single-word message",
			input: "yo",
			want:  "yo" + messageSpaceSuffix,
		},
		{
			desc:  "two-word message",
			input: "yo sup",
			want:  "yo  sup",
		},
		{
			desc:  "multi-word message",
			input: "yo sup man",
			want:  "yo  sup man",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if got := bypassSameMessageDetection(tc.input); got != tc.want {
				t.Errorf("bypassSameMessageDetection() = %q, want %q", got, tc.want)
			}
		})
	}
}

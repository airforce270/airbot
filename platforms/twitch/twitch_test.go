package twitch

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/airforce270/airbot/apiclients/twitchtest"
	"github.com/airforce270/airbot/database/databasetest"
	"github.com/google/go-cmp/cmp"
)

func TestTwitch_CurrentUsers(t *testing.T) {
	t.Parallel()
	db := databasetest.New(t)
	server := newTestServer()
	tw := NewForTesting(server.URL, db)

	got, err := tw.CurrentUsers()
	if err != nil {
		t.Fatalf("CurrentUsers unexpected error: %v", err)
	}

	want := []string{"user1", "user2"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("CurrentUsers() diff (-want +got):\n%s", diff)
	}
}

func TestLowercaseAll(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			if got := bypassSameMessageDetection(tc.input); got != tc.want {
				t.Errorf("bypassSameMessageDetection() = %q, want %q", got, tc.want)
			}
		})
	}
}

func newTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/chat/chatters") {
			fmt.Fprint(w, twitchtest.GetChannelChatChattersResp)
		} else {
			log.Printf("Unknown URL sent to test server: %s", r.URL.Path)
		}
	}))
}

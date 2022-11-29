package twitchtmi

import (
	"fmt"
	"testing"

	"github.com/airforce270/airbot/apiclients/twitchtmi/twitchtmitest"
	"github.com/airforce270/airbot/testing/fakeserver"

	"github.com/google/go-cmp/cmp"
)

var (
	originalBaseURL = BaseURL
)

func TestFetchChatters(t *testing.T) {
	server := fakeserver.New()
	server.AddOnClose(func() { originalBaseURL = BaseURL })
	defer server.Close()
	BaseURL = server.URL()

	tests := []struct {
		useResp string
		want    *FetchChattersResponse
	}{
		{
			useResp: twitchtmitest.FetchChattersManyChattersResp,
			want: &FetchChattersResponse{
				ChatterCount: 15,
				Chatters: Chatters{
					Broadcaster: []string{"airforce2700"},
					VIPs:        []string{},
					Moderators:  []string{"af2bot", "streamelements", "fossabot", "ip0g"},
					Staff:       []string{},
					Admins:      []string{},
					GlobalMods:  []string{},
					Viewers:     []string{"augustcelery", "bapplesas", "dafke_", "ellagarten", "esattt", "femboynv", "givemeanonion", "iizzybeth", "iqkev", "rockn__"},
				},
			},
		},
	}

	for _, tc := range tests {
		server.Resp = tc.useResp
		t.Run(fmt.Sprintf("%d", tc.want.ChatterCount), func(t *testing.T) {
			got, err := FetchChatters("user1")
			if err != nil {
				t.Fatalf("FetchChatters() unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("FetchChatters() diff (-want +got):\n%s", diff)
			}
		})
		server.Reset()
	}
}

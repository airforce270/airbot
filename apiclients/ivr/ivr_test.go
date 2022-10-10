package ivr

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

var (
	originalIvrBaseUrl = ivrBaseURL

	ivrTwitchUserNotStreamingResp = `{"banned":false,"displayName":"xQc","login":"xqc","id":"71092938","bio":"THE BEST AT ABSOLUTELY EVERYTHING. THE JUICER. LEADER OF THE JUICERS.","follows":207,"followers":11226373,"profileViewCount":524730962,"chatColor":"#0000FF","logo":"https://static-cdn.jtvnw.net/jtv_user_pictures/xqc-profile_image-9298dca608632101-600x600.jpeg","banner":"https://static-cdn.jtvnw.net/jtv_user_pictures/83e86af1-9a6c-42b1-98e2-3f6238a744b5-profile_banner-480.png","verifiedBot":false,"createdAt":"2014-09-12T23:50:05.989719Z","updatedAt":"2022-10-06T20:43:00.256907Z","emotePrefix":"xqc","roles":{"isAffiliate":false,"isPartner":true,"isStaff":null},"badges":[{"setID":"partner","title":"Verified","description":"Verified","version":"1"}],"chatSettings":{"chatDelayMs":0,"followersOnlyDurationMinutes":1440,"slowModeDurationSeconds":null,"blockLinks":false,"isSubscribersOnlyModeEnabled":false,"isEmoteOnlyModeEnabled":false,"isFastSubsModeEnabled":false,"isUniqueChatModeEnabled":false,"requireVerifiedAccount":false,"rules":["English please","Fresh memes"]},"stream":null,"lastBroadcast":{"startedAt":"2022-10-06T22:47:39.840638Z","title":"🟧JUICED EP2. !FANSLY🟧CLICK NOW🟧FT. JERMA🟧& AUSTIN🟧& LUDWIG🟧& CONNOREATSPANTS🟧& ME🟧JOIN NOW🟧FAST🟧BEFORE I LOSE IT🟧BIG🟧#SPONSORED"},"panels":[{"id":"124112525"},{"id":"98109996"},{"id":"44997828"},{"id":"32221884"},{"id":"12592823"},{"id":"6720150"},{"id":"77693957"},{"id":"12592818"},{"id":"8847001"},{"id":"22113669"},{"id":"8847029"},{"id":"22360616"},{"id":"14506832"},{"id":"22360618"}]}`
	ivrTwitchUserStreamingResp    = `{"banned":false,"displayName":"xQt0001","login":"xqt0001","id":"591140996","bio":"GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA ","follows":144,"followers":1940,"profileViewCount":190,"chatColor":"#FDFF00","logo":"https://static-cdn.jtvnw.net/jtv_user_pictures/a5fecb44-bb38-4739-aa04-371cc1ea4152-profile_image-600x600.png","banner":"https://static-cdn.jtvnw.net/jtv_user_pictures/fb78db5c-0078-4e85-bfe1-f527f19d9a22-profile_banner-480.jpeg","verifiedBot":false,"createdAt":"2020-10-02T16:25:47.819212Z","updatedAt":"2022-10-09T03:41:23.622605Z","emotePrefix":"xqt000","roles":{"isAffiliate":true,"isPartner":false,"isStaff":null},"badges":[{"setID":"premium","title":"Prime Gaming","description":"Prime Gaming","version":"1"}],"chatSettings":{"chatDelayMs":0,"followersOnlyDurationMinutes":null,"slowModeDurationSeconds":null,"blockLinks":false,"isSubscribersOnlyModeEnabled":false,"isEmoteOnlyModeEnabled":false,"isFastSubsModeEnabled":false,"isUniqueChatModeEnabled":false,"requireVerifiedAccount":false,"rules":[]},"stream":{"title":"tiktok esport #228 i guess","id":"39929884600","createdAt":"2022-10-09T22:00:33Z","type":"live","viewersCount":77,"game":{"displayName":"Just Chatting"}},"lastBroadcast":{"startedAt":"2022-10-09T22:00:37.637909Z","title":"tiktok esport #228 i guess"},"panels":[]}`
	ivrTwitchUserBannedResp       = `{"banned":true,"banReason":"TOS_INDEFINITE","displayName":"SeaGrade","login":"seagrade","id":"245821818","bio":"unbanned","follows":5,"followers":0,"profileViewCount":150,"chatColor":"#00EBFF","logo":"https://static-cdn.jtvnw.net/jtv_user_pictures/2cbd05a5-5502-4e42-924b-f889dc2221f7-profile_image-600x600.png","banner":"https://static-cdn.jtvnw.net/jtv_user_pictures/f6d8baf2-f2bf-4098-a4b7-f9945bd42ff7-profile_banner-480.jpeg","verifiedBot":false,"createdAt":"2018-08-05T23:43:51.848531Z","updatedAt":"2022-10-04T22:24:03.192561Z","emotePrefix":"","roles":{"isAffiliate":false,"isPartner":false,"isStaff":null},"badges":[{"setID":"glhf-pledge","title":"GLHF Pledge","description":"Signed the GLHF pledge in support for inclusive gaming communities","version":"1"}],"chatSettings":{"chatDelayMs":0,"followersOnlyDurationMinutes":null,"slowModeDurationSeconds":null,"blockLinks":true,"isSubscribersOnlyModeEnabled":false,"isEmoteOnlyModeEnabled":false,"isFastSubsModeEnabled":false,"isUniqueChatModeEnabled":false,"requireVerifiedAccount":true,"rules":[]},"stream":null,"lastBroadcast":{"startedAt":"2018-09-02T23:43:41.435181Z","title":"OBS TEST"},"panels":[{"id":"88030436"}]}`

	ivrTwitchUserNotVerifiedBotResp = `{"banned":false,"displayName":"xQc","login":"xqc","id":"71092938","bio":"THE BEST AT ABSOLUTELY EVERYTHING. THE JUICER. LEADER OF THE JUICERS.","follows":207,"followers":11226373,"profileViewCount":524730962,"chatColor":"#0000FF","logo":"https://static-cdn.jtvnw.net/jtv_user_pictures/xqc-profile_image-9298dca608632101-600x600.jpeg","banner":"https://static-cdn.jtvnw.net/jtv_user_pictures/83e86af1-9a6c-42b1-98e2-3f6238a744b5-profile_banner-480.png","verifiedBot":false,"createdAt":"2014-09-12T23:50:05.989719Z","updatedAt":"2022-10-06T20:43:00.256907Z","emotePrefix":"xqc","roles":{"isAffiliate":false,"isPartner":true,"isStaff":null},"badges":[{"setID":"partner","title":"Verified","description":"Verified","version":"1"}],"chatSettings":{"chatDelayMs":0,"followersOnlyDurationMinutes":1440,"slowModeDurationSeconds":null,"blockLinks":false,"isSubscribersOnlyModeEnabled":false,"isEmoteOnlyModeEnabled":false,"isFastSubsModeEnabled":false,"isUniqueChatModeEnabled":false,"requireVerifiedAccount":false,"rules":["English please","Fresh memes"]},"stream":null,"lastBroadcast":{"startedAt":"2022-10-06T22:47:39.840638Z","title":"🟧JUICED EP2. !FANSLY🟧CLICK NOW🟧FT. JERMA🟧& AUSTIN🟧& LUDWIG🟧& CONNOREATSPANTS🟧& ME🟧JOIN NOW🟧FAST🟧BEFORE I LOSE IT🟧BIG🟧#SPONSORED"},"panels":[{"id":"124112525"},{"id":"98109996"},{"id":"44997828"},{"id":"32221884"},{"id":"12592823"},{"id":"6720150"},{"id":"77693957"},{"id":"12592818"},{"id":"8847001"},{"id":"22113669"},{"id":"8847029"},{"id":"22360616"},{"id":"14506832"},{"id":"22360618"}]}`
	ivrTwitchUserVerifiedBotResp    = `{"banned":false,"displayName":"iP0G","login":"ip0g","id":"429509069","bio":"very p0g","follows":136,"followers":63,"profileViewCount":658,"chatColor":"#00FFFF","logo":"https://static-cdn.jtvnw.net/jtv_user_pictures/dd669686-7694-418f-bd83-5ad418b5bb3b-profile_image-600x600.png","banner":null,"verifiedBot":true,"createdAt":"2019-04-12T05:34:52.280629Z","updatedAt":"2022-09-16T00:10:21.56657Z","emotePrefix":"","roles":{"isAffiliate":false,"isPartner":false,"isStaff":null},"badges":[{"setID":"game-developer","title":"Game Developer","description":"Game Developer for:","version":"1"}],"chatSettings":{"chatDelayMs":0,"followersOnlyDurationMinutes":0,"slowModeDurationSeconds":null,"blockLinks":false,"isSubscribersOnlyModeEnabled":false,"isEmoteOnlyModeEnabled":false,"isFastSubsModeEnabled":false,"isUniqueChatModeEnabled":false,"requireVerifiedAccount":false,"rules":[]},"stream":null,"lastBroadcast":{"startedAt":"2022-03-04T05:45:02.37469Z","title":null},"panels":[{"id":"123839619"}]}`
)

type fakeServer struct {
	s *httptest.Server

	Resp string
}

func (s *fakeServer) Init() {
	ivrBaseURL = s.s.URL
}

func (s *fakeServer) Close() {
	s.s.Close()
	ivrBaseURL = originalIvrBaseUrl
}

func (s *fakeServer) Reset() {
	s.Resp = ""
}

func newFakeServer() *fakeServer {
	s := fakeServer{Resp: ""}
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, s.Resp)
	}))
	s.s = httpServer
	return &s
}

func TestFetchUser(t *testing.T) {
	server := newFakeServer()
	server.Init()
	defer server.Close()

	tests := []struct {
		desc    string
		useResp string
		want    *ivrTwitchUserResponse
	}{
		{
			desc:    "non streaming user",
			useResp: ivrTwitchUserNotStreamingResp,
			want: &ivrTwitchUserResponse{
				IsBanned:          false,
				BanReason:         "",
				DisplayName:       "xQc",
				Login:             "xqc",
				ID:                "71092938",
				Bio:               "THE BEST AT ABSOLUTELY EVERYTHING. THE JUICER. LEADER OF THE JUICERS.",
				FollowCount:       207,
				FollowersCount:    11226373,
				ProfileViewCount:  524730962,
				ChatColor:         "#0000FF",
				ProfilePictureURL: "https://static-cdn.jtvnw.net/jtv_user_pictures/xqc-profile_image-9298dca608632101-600x600.jpeg",
				BannerURL:         "https://static-cdn.jtvnw.net/jtv_user_pictures/83e86af1-9a6c-42b1-98e2-3f6238a744b5-profile_banner-480.png",
				IsVerifiedBot:     false,
				CreatedAt:         time.Date(2014, 9, 12, 23, 50, 5, 989719000, time.UTC),
				UpdatedAt:         time.Date(2022, 10, 6, 20, 43, 0, 256907000, time.UTC),
				EmotePrefix:       "xqc",
				Roles: rolesInfo{
					IsAffiliate: false,
					IsPartner:   true,
					IsStaff:     false,
				},
				Badges: []badgeInfo{
					{
						Set:         "partner",
						Title:       "Verified",
						Description: "Verified",
						Version:     "1",
					},
				},
				ChatSettings: chatSettingsInfo{
					ChatDelayMs:                  0,
					FollowersOnlyDurationMinutes: 1440,
					SlowModeDurationSeconds:      0,
					BlockLinks:                   false,
					IsSubscribersOnlyModeEnabled: false,
					IsEmoteOnlyModeEnabled:       false,
					IsFastSubsModeEnabled:        false,
					IsUniqueChatModeEnabled:      false,
					RequireVerifiedAccount:       false,
					Rules: []string{
						"English please",
						"Fresh memes",
					},
				},
				Stream: nil,
				LastBroadcast: lastBroadcastInfo{
					StartTime: time.Date(2022, 10, 6, 22, 47, 39, 840638000, time.UTC),
					Title:     "🟧JUICED EP2. !FANSLY🟧CLICK NOW🟧FT. JERMA🟧& AUSTIN🟧& LUDWIG🟧& CONNOREATSPANTS🟧& ME🟧JOIN NOW🟧FAST🟧BEFORE I LOSE IT🟧BIG🟧#SPONSORED",
				},
				Panels: []panelInfo{
					{ID: "124112525"},
					{ID: "98109996"},
					{ID: "44997828"},
					{ID: "32221884"},
					{ID: "12592823"},
					{ID: "6720150"},
					{ID: "77693957"},
					{ID: "12592818"},
					{ID: "8847001"},
					{ID: "22113669"},
					{ID: "8847029"},
					{ID: "22360616"},
					{ID: "14506832"},
					{ID: "22360618"},
				},
			},
		},
		{
			desc:    "streaming user",
			useResp: ivrTwitchUserStreamingResp,
			want: &ivrTwitchUserResponse{
				IsBanned:          false,
				BanReason:         "",
				DisplayName:       "xQt0001",
				Login:             "xqt0001",
				ID:                "591140996",
				Bio:               "GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA GAMBA ",
				FollowCount:       144,
				FollowersCount:    1940,
				ProfileViewCount:  190,
				ChatColor:         "#FDFF00",
				ProfilePictureURL: "https://static-cdn.jtvnw.net/jtv_user_pictures/a5fecb44-bb38-4739-aa04-371cc1ea4152-profile_image-600x600.png",
				BannerURL:         "https://static-cdn.jtvnw.net/jtv_user_pictures/fb78db5c-0078-4e85-bfe1-f527f19d9a22-profile_banner-480.jpeg",
				IsVerifiedBot:     false,
				CreatedAt:         time.Date(2020, 10, 2, 16, 25, 47, 819212000, time.UTC),
				UpdatedAt:         time.Date(2022, 10, 9, 3, 41, 23, 622605000, time.UTC),
				EmotePrefix:       "xqt000",
				Roles: rolesInfo{
					IsAffiliate: true,
					IsPartner:   false,
					IsStaff:     false,
				},
				Badges: []badgeInfo{
					{
						Set:         "premium",
						Title:       "Prime Gaming",
						Description: "Prime Gaming",
						Version:     "1",
					},
				},
				ChatSettings: chatSettingsInfo{
					ChatDelayMs:                  0,
					FollowersOnlyDurationMinutes: 0,
					SlowModeDurationSeconds:      0,
					BlockLinks:                   false,
					IsSubscribersOnlyModeEnabled: false,
					IsEmoteOnlyModeEnabled:       false,
					IsFastSubsModeEnabled:        false,
					IsUniqueChatModeEnabled:      false,
					RequireVerifiedAccount:       false,
					Rules:                        []string{},
				},
				Stream: &streamInfo{
					Title:        "tiktok esport #228 i guess",
					ID:           "39929884600",
					StartTime:    time.Date(2022, 10, 9, 22, 0, 33, 0, time.UTC),
					Type:         "live",
					ViewersCount: 77,
					Game:         gameInfo{DisplayName: "Just Chatting"},
				},
				LastBroadcast: lastBroadcastInfo{
					StartTime: time.Date(2022, 10, 9, 22, 0, 37, 637909000, time.UTC),
					Title:     "tiktok esport #228 i guess",
				},
				Panels: []panelInfo{},
			},
		},
		{
			desc:    "banned user",
			useResp: ivrTwitchUserBannedResp,
			want: &ivrTwitchUserResponse{
				IsBanned:          true,
				BanReason:         "TOS_INDEFINITE",
				DisplayName:       "SeaGrade",
				Login:             "seagrade",
				ID:                "245821818",
				Bio:               "unbanned",
				FollowCount:       5,
				FollowersCount:    0,
				ProfileViewCount:  150,
				ChatColor:         "#00EBFF",
				ProfilePictureURL: "https://static-cdn.jtvnw.net/jtv_user_pictures/2cbd05a5-5502-4e42-924b-f889dc2221f7-profile_image-600x600.png",
				BannerURL:         "https://static-cdn.jtvnw.net/jtv_user_pictures/f6d8baf2-f2bf-4098-a4b7-f9945bd42ff7-profile_banner-480.jpeg",
				IsVerifiedBot:     false,
				CreatedAt:         time.Date(2018, 8, 5, 23, 43, 51, 848531000, time.UTC),
				UpdatedAt:         time.Date(2022, 10, 4, 22, 24, 3, 192561000, time.UTC),
				EmotePrefix:       "",
				Roles: rolesInfo{
					IsAffiliate: false,
					IsPartner:   false,
					IsStaff:     false,
				},
				Badges: []badgeInfo{
					{
						Set:         "glhf-pledge",
						Title:       "GLHF Pledge",
						Description: "Signed the GLHF pledge in support for inclusive gaming communities",
						Version:     "1",
					},
				},
				ChatSettings: chatSettingsInfo{
					ChatDelayMs:                  0,
					FollowersOnlyDurationMinutes: 0,
					SlowModeDurationSeconds:      0,
					BlockLinks:                   true,
					IsSubscribersOnlyModeEnabled: false,
					IsEmoteOnlyModeEnabled:       false,
					IsFastSubsModeEnabled:        false,
					IsUniqueChatModeEnabled:      false,
					RequireVerifiedAccount:       true,
					Rules:                        []string{},
				},
				Stream: nil,
				LastBroadcast: lastBroadcastInfo{
					StartTime: time.Date(2018, 9, 2, 23, 43, 41, 435181000, time.UTC),
					Title:     "OBS TEST",
				},
				Panels: []panelInfo{
					{ID: "88030436"},
				},
			},
		},
	}

	for _, tc := range tests {
		server.Resp = tc.useResp
		t.Run(tc.desc, func(t *testing.T) {
			got, err := FetchUser("fake-username")
			if err != nil {
				t.Fatalf("FetchUser() unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("FetchUser() diff (-want +got):\n%s", diff)
			}
		})
		server.Reset()
	}
}

func TestIsVerifiedBot(t *testing.T) {
	server := newFakeServer()
	server.Init()
	defer server.Close()

	tests := []struct {
		desc    string
		useResp string
		want    bool
	}{
		{
			desc:    "not verified bot",
			useResp: ivrTwitchUserNotVerifiedBotResp,
			want:    false,
		},
		{
			desc:    "verified bot",
			useResp: ivrTwitchUserVerifiedBotResp,
			want:    true,
		},
	}

	for _, tc := range tests {
		server.Resp = tc.useResp
		t.Run(tc.desc, func(t *testing.T) {
			user, err := FetchUser("fake-username")
			if err != nil {
				t.Fatalf("IsVerifiedBot() unexpected error: %v", err)
			}

			if got := user.IsVerifiedBot; got != tc.want {
				t.Errorf("IsVerifiedBot() = %t, want %t", got, tc.want)
			}
		})
		server.Reset()
	}
}

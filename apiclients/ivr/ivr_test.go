package ivr

import (
	"testing"
	"time"

	"github.com/airforce270/airbot/apiclients/ivr/ivrtest"
	"github.com/airforce270/airbot/testing/fakeserver"

	"github.com/google/go-cmp/cmp"
)

var (
	originalIvrBaseUrl = BaseURL
)

func TestFetchUser(t *testing.T) {
	server := fakeserver.New()
	server.AddOnClose(func() { originalIvrBaseUrl = BaseURL })
	defer server.Close()
	BaseURL = server.URL()

	tests := []struct {
		desc    string
		useResp string
		want    []*TwitchUsersResponseItem
	}{
		{
			desc:    "non streaming user",
			useResp: ivrtest.TwitchUsersNotStreamingResp,
			want: []*TwitchUsersResponseItem{
				{
					IsBanned:          false,
					BanReason:         "",
					DisplayName:       "xQc",
					Username:          "xqc",
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
		},
		{
			desc:    "streaming user",
			useResp: ivrtest.TwitchUsersStreamingResp,
			want: []*TwitchUsersResponseItem{
				{
					IsBanned:          false,
					BanReason:         "",
					DisplayName:       "xQt0001",
					Username:          "xqt0001",
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
		},
		{
			desc:    "banned user",
			useResp: ivrtest.TwitchUsersBannedResp,
			want: []*TwitchUsersResponseItem{
				{
					IsBanned:          true,
					BanReason:         "TOS_INDEFINITE",
					DisplayName:       "SeaGrade",
					Username:          "seagrade",
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
		},
	}

	for _, tc := range tests {
		server.Resp = tc.useResp
		t.Run(tc.desc, func(t *testing.T) {
			got, err := FetchUsers("fake-username")
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

func TestFetchModsAndVIPs(t *testing.T) {
	server := fakeserver.New()
	server.AddOnClose(func() { originalIvrBaseUrl = BaseURL })
	defer server.Close()
	BaseURL = server.URL()

	tests := []struct {
		desc    string
		useResp string
		want    *ModsAndVIPsResponse
	}{
		{
			desc:    "no mods or vips",
			useResp: ivrtest.ModsAndVIPsNoneResp,
			want: &ModsAndVIPsResponse{
				Mods: []*ModOrVIPUser{},
				VIPs: []*ModOrVIPUser{},
			},
		},
		{
			desc:    "mods only",
			useResp: ivrtest.ModsAndVIPsModsOnlyResp,
			want: &ModsAndVIPsResponse{
				Mods: []*ModOrVIPUser{
					{
						ID:          "429509069",
						Username:    "ip0g",
						DisplayName: "iP0G",
						GrantedAt:   time.Date(2022, 10, 3, 19, 55, 0, 137915435, time.UTC),
					},
					{
						ID:          "834890604",
						Username:    "af2bot",
						DisplayName: "af2bot",
						GrantedAt:   time.Date(2022, 10, 9, 8, 13, 17, 829797513, time.UTC),
					},
				},
				VIPs: []*ModOrVIPUser{},
			},
		},
		{
			desc:    "large, many mods and vips",
			useResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: &ModsAndVIPsResponse{
				Mods: []*ModOrVIPUser{
					{
						ID:          "100135110",
						Username:    "streamelements",
						DisplayName: "StreamElements",
						GrantedAt:   time.Date(2018, 7, 24, 8, 29, 21, 757709759, time.UTC),
					},
					{
						ID:          "237719657",
						Username:    "fossabot",
						DisplayName: "Fossabot",
						GrantedAt:   time.Date(2020, 8, 16, 20, 51, 55, 198556309, time.UTC),
					},
					{
						ID:          "191202519",
						Username:    "spintto",
						DisplayName: "spintto",
						GrantedAt:   time.Date(2022, 3, 8, 14, 59, 43, 671830635, time.UTC),
					},
					{
						ID:          "514751411",
						Username:    "hnoace",
						DisplayName: "HNoAce",
						GrantedAt:   time.Date(2022, 8, 9, 13, 35, 14, 995445410, time.UTC),
					},
				},
				VIPs: []*ModOrVIPUser{
					{
						ID:          "150790620",
						Username:    "bakonsword",
						DisplayName: "bakonsword",
						GrantedAt:   time.Date(2022, 2, 20, 19, 39, 12, 355546493, time.UTC),
					},
					{
						ID:          "145484970",
						Username:    "alyjiaht_t",
						DisplayName: "alyjiahT_T",
						GrantedAt:   time.Date(2022, 2, 25, 5, 42, 16, 48233372, time.UTC),
					},
					{
						ID:          "205748697",
						Username:    "avbest",
						DisplayName: "AVBest",
						GrantedAt:   time.Date(2022, 3, 8, 14, 31, 49, 869620222, time.UTC),
					},
					{
						ID:          "69184756",
						Username:    "zaintew_",
						DisplayName: "Zaintew_",
						GrantedAt:   time.Date(2022, 9, 17, 21, 43, 57, 737612548, time.UTC),
					},
					{
						ID:          "505131195",
						Username:    "captkayy",
						DisplayName: "captkayy",
						GrantedAt:   time.Date(2022, 9, 25, 20, 15, 59, 332859708, time.UTC),
					},
					{
						ID:          "425925187",
						Username:    "seagrad",
						DisplayName: "seagrad",
						GrantedAt:   time.Date(2022, 10, 5, 5, 51, 51, 432004125, time.UTC),
					},
					{
						ID:          "222316577",
						Username:    "dafkeee",
						DisplayName: "Dafkeee",
						GrantedAt:   time.Date(2022, 10, 5, 5, 52, 2, 130647633, time.UTC),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		server.Resp = tc.useResp
		t.Run(tc.desc, func(t *testing.T) {
			got, err := FetchModsAndVIPs("fakeusername")
			if err != nil {
				t.Fatalf("FetchModsAndVIPs() unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("FetchModsAndVIPs() diff (-want +got):\n%s", diff)
			}
		})
		server.Reset()
	}
}

func TestFetchFounders(t *testing.T) {
	server := fakeserver.New()
	server.AddOnClose(func() { originalIvrBaseUrl = BaseURL })
	defer server.Close()
	BaseURL = server.URL()

	tests := []struct {
		desc    string
		useResp string
		want    *FoundersResponse
	}{
		{
			desc:    "no founders 404",
			useResp: ivrtest.FoundersNone404Resp,
			want: &FoundersResponse{
				Founders: nil,
			},
		},
		{
			desc:    "no founders",
			useResp: ivrtest.FoundersNoneResp,
			want: &FoundersResponse{
				Founders: []*Founder{},
			},
		},
		{
			desc:    "founders",
			useResp: ivrtest.FoundersNormalResp,
			want: &FoundersResponse{
				Founders: []*Founder{
					{
						ID:                "415575292",
						Username:          "fishyykingyy",
						DisplayName:       "FishyyKingyy",
						InitiallySubbedAt: time.Date(2022, 7, 31, 0, 41, 6, 0, time.UTC),
					},
					{
						ID:                "267287250",
						Username:          "eljulidi1337",
						DisplayName:       "eljulidi1337",
						InitiallySubbedAt: time.Date(2022, 8, 13, 19, 46, 18, 0, time.UTC),
					},
					{
						ID:                "89075062",
						Username:          "sammist",
						DisplayName:       "SamMist",
						InitiallySubbedAt: time.Date(2022, 8, 16, 15, 24, 49, 0, time.UTC),
						IsSubscribed:      true,
					},
					{
						ID:                "190634299",
						Username:          "leochansz",
						DisplayName:       "Leochansz",
						InitiallySubbedAt: time.Date(2022, 8, 16, 15, 41, 52, 0, time.UTC),
						IsSubscribed:      true,
					},
					{
						ID:                "143232353",
						Username:          "lexieuzumaki7",
						DisplayName:       "lexieuzumaki7",
						InitiallySubbedAt: time.Date(2022, 8, 17, 5, 7, 54, 0, time.UTC),
					},
					{
						ID:                "65602310",
						Username:          "contravz",
						DisplayName:       "ContraVz",
						InitiallySubbedAt: time.Date(2022, 8, 17, 21, 44, 28, 0, time.UTC),
					},
					{
						ID:                "232875294",
						Username:          "rott______",
						DisplayName:       "rott______",
						InitiallySubbedAt: time.Date(2022, 8, 18, 0, 41, 48, 0, time.UTC),
						IsSubscribed:      true,
					},
					{
						ID:                "610912094",
						Username:          "dankjuicer",
						DisplayName:       "DankJuicer",
						InitiallySubbedAt: time.Date(2022, 8, 18, 0, 48, 10, 0, time.UTC),
					},
					{
						ID:                "671024739",
						Username:          "kronikz____",
						DisplayName:       "kronikZ____",
						InitiallySubbedAt: time.Date(2022, 8, 20, 20, 39, 11, 0, time.UTC),
					},
					{
						ID:                "408538669",
						Username:          "blemplob",
						DisplayName:       "blemplob",
						InitiallySubbedAt: time.Date(2022, 8, 24, 1, 48, 53, 0, time.UTC),
						IsSubscribed:      true,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		server.Resp = tc.useResp
		t.Run(tc.desc, func(t *testing.T) {
			got, err := FetchFounders("fakeusername")
			if err != nil {
				t.Fatalf("FetchFounders() unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("FetchFounders() diff (-want +got):\n%s", diff)
			}
		})
		server.Reset()
	}
}

func TestIsVerifiedBot(t *testing.T) {
	server := fakeserver.New()
	server.AddOnClose(func() { originalIvrBaseUrl = BaseURL })
	defer server.Close()
	BaseURL = server.URL()

	tests := []struct {
		desc    string
		useResp string
		want    bool
	}{
		{
			desc:    "not verified bot",
			useResp: ivrtest.TwitchUsersNotVerifiedBotResp,
			want:    false,
		},
		{
			desc:    "verified bot",
			useResp: ivrtest.TwitchUsersVerifiedBotResp,
			want:    true,
		},
	}

	for _, tc := range tests {
		server.Resp = tc.useResp
		t.Run(tc.desc, func(t *testing.T) {
			users, err := FetchUsers("fake-username")
			if err != nil {
				t.Fatalf("IsVerifiedBot() unexpected error: %v", err)
			}
			if len(users) != 1 {
				t.Fatalf("IsVerifiedBot() len(users) == %d, want 1", len(users))
			}

			if got := users[0].IsVerifiedBot; got != tc.want {
				t.Errorf("IsVerifiedBot() = %t, want %t", got, tc.want)
			}
		})
		server.Reset()
	}
}

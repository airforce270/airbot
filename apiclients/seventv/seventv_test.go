package seventv_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/airforce270/airbot/apiclients/seventv"
	"github.com/airforce270/airbot/apiclients/seventv/seventvtest"
	"github.com/airforce270/airbot/testing/fakeserver"

	"github.com/google/go-cmp/cmp"
)

func TestAddEmote(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc    string
		useResp string
		wantErr error
	}{
		{
			desc:    "success",
			useResp: seventvtest.MutateEmoteSuccessResp,
			wantErr: nil,
		},
		{
			desc:    "not authorized",
			useResp: seventvtest.MutateEmoteNotAuthorizedResp,
			wantErr: seventv.ErrNotAuthorized,
		},
		{
			desc:    "already added",
			useResp: seventvtest.MutateEmoteAlreadyExistsResp,
			wantErr: seventv.ErrEmoteAlreadyEnabled,
		},
		{
			desc:    "id not found",
			useResp: seventvtest.MutateEmoteIDNotFoundResp,
			wantErr: seventv.ErrEmoteNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			server := fakeserver.New()
			defer server.Close()

			server.Resps = []string{tc.useResp}
			client := seventv.NewClient(ctx, *server.URL(t), "" /* accessToken */)

			err := client.AddEmote(ctx, "fake-emote-set-1", "fake-emote-1")
			if err != nil && tc.wantErr == nil {
				t.Fatalf("AddEmote() unexpected error: %v", err)
			}
			if err == nil && tc.wantErr != nil {
				t.Fatal("AddEmote() had no err, but expected one")
			}
			if err != nil && tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("AddEmote() expected err %v to be %v", err, tc.wantErr)
				}
			}

			const wantBody = `{"query":` +
				`"mutation ($action:ListItemAction!$emote_id:ObjectID!$emote_set_id:ObjectID!){` +
				`emoteSet(id: $emote_set_id){id,emotes(id: $emote_id, action: $action){id,name}}` +
				`}",` +
				`"variables":{"action":"ADD","emote_id":"fake-emote-1","emote_set_id":"fake-emote-set-1"}` +
				"}\n"

			gotBody, err := io.ReadAll(server.Reqs[0].Body)
			if err != nil {
				t.Fatalf("AddEmote() failed to read req body: %v", err)
			}
			if diff := cmp.Diff(wantBody, string(gotBody)); diff != "" {
				t.Errorf("AddEmote() req body diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAddEmoteWithAlias(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc    string
		useResp string
		wantErr error
	}{
		{
			desc:    "success",
			useResp: seventvtest.MutateEmoteSuccessResp,
			wantErr: nil,
		},
		{
			desc:    "not authorized",
			useResp: seventvtest.MutateEmoteNotAuthorizedResp,
			wantErr: seventv.ErrNotAuthorized,
		},
		{
			desc:    "already added",
			useResp: seventvtest.MutateEmoteAlreadyExistsResp,
			wantErr: seventv.ErrEmoteAlreadyEnabled,
		},
		{
			desc:    "id not found",
			useResp: seventvtest.MutateEmoteIDNotFoundResp,
			wantErr: seventv.ErrEmoteNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			server := fakeserver.New()
			defer server.Close()

			server.Resps = []string{tc.useResp}
			client := seventv.NewClient(ctx, *server.URL(t), "" /* accessToken */)

			err := client.AddEmoteWithAlias(ctx, "fake-emote-set-1", "fake-emote-1", "fake-alias")
			if err != nil && tc.wantErr == nil {
				t.Fatalf("AddEmoteWithAlias() unexpected error: %v", err)
			}
			if err == nil && tc.wantErr != nil {
				t.Fatal("AddEmoteWithAlias() had no err, but expected one")
			}
			if err != nil && tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("AddEmoteWithAlias() expected err %v to be %v", err, tc.wantErr)
				}
			}

			const wantBody = `{"query":` +
				`"mutation ($action:ListItemAction!$emote_id:ObjectID!$emote_set_id:ObjectID!$name:String!){` +
				`emoteSet(id: $emote_set_id){id,emotes(id: $emote_id, action: $action, name: $name){id,name}}` +
				`}",` +
				`"variables":{"action":"ADD","emote_id":"fake-emote-1","emote_set_id":"fake-emote-set-1","name":"fake-alias"}` +
				"}\n"

			gotBody, err := io.ReadAll(server.Reqs[0].Body)
			if err != nil {
				t.Fatalf("AddEmoteWithAlias() failed to read req body: %v", err)
			}
			if diff := cmp.Diff(wantBody, string(gotBody)); diff != "" {
				t.Errorf("AddEmoteWithAlias() req body diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFetchUserConnectionByTwitchUserId(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc    string
		useResp string
		want    *seventv.PlatformConnection
	}{
		{
			desc:    "large live channel",
			useResp: seventvtest.FetchUserConnectionByTwitchUserIdSmallNonSubResp,
			want: &seventv.PlatformConnection{
				Platform:      "TWITCH",
				ID:            "181950834",
				Username:      "airforce2700",
				DisplayName:   "airforce2700",
				LinkedAt:      seventv.UnixTimeMs(time.Date(2022, 3, 2, 6, 50, 30, 0, time.UTC)),
				EmoteCapacity: 1000,
				EmoteSet: seventv.EmoteSet{
					ID:   "621f13b614f489808df5d58e",
					Name: "airforce2700's Emotes",
					Tags: []string{},
					Emotes: []seventv.Emote{
						{
							ID:         "63f9c34b04e4a9fd8ee1c581",
							Name:       "kok",
							UpdateTime: seventv.UnixTimeMs(time.Date(2023, 10, 12, 2, 40, 35, 948000000, time.UTC)),
							Creator:    "621f13b614f489808df5d58e",
							Data: seventv.EmoteData{
								ID:        "63f9c34b04e4a9fd8ee1c581",
								Name:      "kok",
								Tags:      []string{"hololive", "zeta", "drawnbychroneco"},
								Lifecycle: 3,
								States:    []string{"LISTED"},
								Listed:    true,
								Animated:  true,
								Owner: seventv.Owner{
									ID:          "6058c630b4d31e459faae649",
									Username:    "vulpeshd",
									DisplayName: "VulpesHD",
									AvatarURL:   "//cdn.7tv.app/user/6058c630b4d31e459faae649/av_6438b54d23e9c459eb14c7ca/3x.webp",
									Style: seventv.Style{
										Paint: 849892095,
									},
									RoleIDs: []string{
										"60724f65e93d828bf8858789",
										"631ef5ea03e9beb96f849a7e",
										"6076a86b09a4c63a38ebe801",
										"62b48deb791a15a25c2a0354",
									},
								},
								Host: seventv.Host{
									BaseURL: "//cdn.7tv.app/emote/63f9c34b04e4a9fd8ee1c581",
									Files: []seventv.File{
										{
											Name:       "1x.avif",
											StaticName: "1x_static.avif",
											Width:      33,
											Height:     32,
											FrameCount: 23,
											Size:       7087,
											Format:     "AVIF",
										},
										{
											Name:       "1x.webp",
											StaticName: "1x_static.webp",
											Width:      33,
											Height:     32,
											FrameCount: 10,
											Size:       10330,
											Format:     "WEBP",
										},
										{
											Name:       "2x.avif",
											StaticName: "2x_static.avif",
											Width:      66,
											Height:     64,
											FrameCount: 24,
											Size:       14469,
											Format:     "AVIF",
										},
										{
											Name:       "2x.webp",
											StaticName: "2x_static.webp",
											Width:      66,
											Height:     64,
											FrameCount: 11,
											Size:       26984,
											Format:     "WEBP",
										},
										{
											Name:       "3x.avif",
											StaticName: "3x_static.avif",
											Width:      99,
											Height:     96,
											FrameCount: 24,
											Size:       22887,
											Format:     "AVIF",
										},
										{
											Name:       "3x.webp",
											StaticName: "3x_static.webp",
											Width:      99,
											Height:     96,
											FrameCount: 17,
											Size:       47054,
											Format:     "WEBP",
										},
										{
											Name:       "4x.avif",
											StaticName: "4x_static.avif",
											Width:      132,
											Height:     128,
											FrameCount: 24,
											Size:       31625,
											Format:     "AVIF",
										},
										{
											Name:       "4x.webp",
											StaticName: "4x_static.webp",
											Width:      132,
											Height:     128,
											FrameCount: 21,
											Size:       72228,
											Format:     "WEBP",
										},
									},
								},
							},
						},
						{
							ID:         "6535d68eaf0fd607b5e8e98f",
							Name:       "librarySecurity",
							UpdateTime: seventv.UnixTimeMs(time.Date(2023, 10, 26, 4, 29, 47, 546000000, time.UTC)),
							Creator:    "621f13b614f489808df5d58e",
							Data: seventv.EmoteData{
								ID:        "6535d68eaf0fd607b5e8e98f",
								Name:      "librarySecurity",
								Lifecycle: 3,
								States:    []string{"LISTED", "NO_PERSONAL"},
								Listed:    true,
								Owner: seventv.Owner{
									ID:          "61cbe6b5ef5a587a0745e707",
									Username:    "ri3zo",
									DisplayName: "ri3zo",
									AvatarURL:   "//cdn.7tv.app/user/61cbe6b5ef5a587a0745e707/av_64cdc9d294a971c9f0e719dc/3x.webp",
									Style: seventv.Style{
										Paint: -5635841,
									},
									RoleIDs: []string{
										"6076a86b09a4c63a38ebe801",
										"62b48deb791a15a25c2a0354",
									},
								},
								Host: seventv.Host{
									BaseURL: "//cdn.7tv.app/emote/6535d68eaf0fd607b5e8e98f",
									Files: []seventv.File{
										{
											Name:       "1x.avif",
											StaticName: "1x_static.avif",
											Width:      32,
											Height:     32,
											FrameCount: 1,
											Size:       1136,
											Format:     "AVIF",
										},
										{
											Name:       "1x.webp",
											StaticName: "1x_static.webp",
											Width:      32,
											Height:     32,
											FrameCount: 1,
											Size:       1362,
											Format:     "WEBP",
										},
										{
											Name:       "2x.avif",
											StaticName: "2x_static.avif",
											Width:      64,
											Height:     64,
											FrameCount: 1,
											Size:       2060,
											Format:     "AVIF",
										},
										{
											Name:       "2x.webp",
											StaticName: "2x_static.webp",
											Width:      64,
											Height:     64,
											FrameCount: 1,
											Size:       3804,
											Format:     "WEBP",
										},
										{
											Name:       "3x.avif",
											StaticName: "3x_static.avif",
											Width:      96,
											Height:     96,
											FrameCount: 1,
											Size:       3043,
											Format:     "AVIF",
										},
										{
											Name:       "3x.webp",
											StaticName: "3x_static.webp",
											Width:      96,
											Height:     96,
											FrameCount: 1,
											Size:       7172,
											Format:     "WEBP",
										},
										{
											Name:       "4x.avif",
											StaticName: "4x_static.avif",
											Width:      128,
											Height:     128,
											FrameCount: 1,
											Size:       4144,
											Format:     "AVIF",
										},
										{
											Name:       "4x.webp",
											StaticName: "4x_static.webp",
											Width:      128,
											Height:     128,
											FrameCount: 1,
											Size:       11024,
											Format:     "WEBP",
										},
									},
								},
							},
						},
						{
							ID:         "654d933aca6c300e60320794",
							Name:       "MacyLookingAtYou",
							UpdateTime: seventv.UnixTimeMs(time.Date(2023, 11, 10, 2, 21, 37, 849000000, time.UTC)),
							Creator:    "621f13b614f489808df5d58e",
							Data: seventv.EmoteData{
								ID:        "654d933aca6c300e60320794",
								Name:      "MacyLookingAtYou",
								Lifecycle: 3,
								States:    []string{"LISTED"},
								Listed:    true,
								Owner: seventv.Owner{
									ID:          "621f13b614f489808df5d58e",
									Username:    "airforce2700",
									DisplayName: "airforce2700",
									AvatarURL:   "//cdn.7tv.app/",
									RoleIDs:     []string{"62b48deb791a15a25c2a0354"},
								},
								Host: seventv.Host{
									BaseURL: "//cdn.7tv.app/emote/654d933aca6c300e60320794",
									Files: []seventv.File{
										{
											Name:       "1x.avif",
											StaticName: "1x_static.avif",
											Width:      32,
											Height:     32,
											FrameCount: 1,
											Size:       752,
											Format:     "AVIF",
										},
										{
											Name:       "1x.webp",
											StaticName: "1x_static.webp",
											Width:      32,
											Height:     32,
											FrameCount: 1,
											Size:       1286,
											Format:     "WEBP",
										},
										{
											Name:       "2x.avif",
											StaticName: "2x_static.avif",
											Width:      64,
											Height:     64,
											FrameCount: 1,
											Size:       1176,
											Format:     "AVIF",
										},
										{
											Name:       "2x.webp",
											StaticName: "2x_static.webp",
											Width:      64,
											Height:     64,
											FrameCount: 1,
											Size:       4104,
											Format:     "WEBP",
										},
										{
											Name:       "3x.avif",
											StaticName: "3x_static.avif",
											Width:      96,
											Height:     96,
											FrameCount: 1,
											Size:       1678,
											Format:     "AVIF",
										},
										{
											Name:       "3x.webp",
											StaticName: "3x_static.webp",
											Width:      96,
											Height:     96,
											FrameCount: 1,
											Size:       7998,
											Format:     "WEBP",
										},
										{
											Name:       "4x.avif",
											StaticName: "4x_static.avif",
											Width:      128,
											Height:     128,
											FrameCount: 1,
											Size:       2261,
											Format:     "AVIF",
										},
										{
											Name:       "4x.webp",
											StaticName: "4x_static.webp",
											Width:      128,
											Height:     128,
											FrameCount: 1,
											Size:       12612,
											Format:     "WEBP",
										},
									},
								},
							},
						},
					},
					EmoteCount: 3,
					Capacity:   1000,
					Owner: seventv.Owner{
						ID:          "621f13b614f489808df5d58e",
						Username:    "airforce2700",
						DisplayName: "airforce2700",
						AvatarURL:   "//cdn.7tv.app/",
						RoleIDs:     []string{"62b48deb791a15a25c2a0354"},
					},
				},
				User: seventv.User{
					ID:          "621f13b614f489808df5d58e",
					Username:    "airforce2700",
					DisplayName: "airforce2700",
					CreateTime:  seventv.UnixTimeMs(time.Date(2022, 3, 2, 6, 50, 30, 0, time.UTC)),
					AvatarURL:   "//cdn.7tv.app/",
					Bio:         "PagMan",
					RoleIDs:     []string{"62b48deb791a15a25c2a0354"},
					Connections: []seventv.PlatformConnection{
						{
							Platform:      "TWITCH",
							ID:            "181950834",
							Username:      "airforce2700",
							DisplayName:   "airforce2700",
							LinkedAt:      seventv.UnixTimeMs(time.Date(2022, 3, 2, 6, 50, 30, 0, time.UTC)),
							EmoteCapacity: 1000,
							EmoteSet:      seventv.EmoteSet{ID: "621f13b614f489808df5d58e", Tags: []string{}},
						},
						{
							Platform:      "KICK",
							ID:            "7426331",
							Username:      "airfors",
							DisplayName:   "airfors",
							LinkedAt:      seventv.UnixTimeMs(time.Date(2023, 6, 20, 3, 30, 9, 511000000, time.UTC)),
							EmoteCapacity: 600,
							EmoteSet:      seventv.EmoteSet{ID: "621f13b614f489808df5d58e", Tags: []string{}},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			server := fakeserver.New()
			defer server.Close()

			server.Resps = []string{tc.useResp}
			client := seventv.NewClient(ctx, *server.URL(t), "" /* accessToken */)

			got, err := client.FetchUserConnectionByTwitchUserId("user1")
			if err != nil {
				t.Fatalf("FetchUserConnectionByTwitchUserId() unexpected error: %v", err)
			}

			cmpOpts := []cmp.Option{seventvtest.Transformer}

			if diff := cmp.Diff(tc.want, got, cmpOpts...); diff != "" {
				t.Errorf("FetchUserConnectionByTwitchUserId() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRemoveEmote(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc    string
		useResp string
		wantErr error
	}{
		{
			desc:    "success",
			useResp: seventvtest.MutateEmoteSuccessResp,
			wantErr: nil,
		},
		{
			desc:    "not authorized",
			useResp: seventvtest.MutateEmoteNotAuthorizedResp,
			wantErr: seventv.ErrNotAuthorized,
		},
		{
			desc:    "not added",
			useResp: seventvtest.MutateEmoteIDNotEnabledResp,
			wantErr: seventv.ErrEmoteNotEnabled,
		},
		{
			desc:    "id not found",
			useResp: seventvtest.MutateEmoteIDNotFoundResp,
			wantErr: seventv.ErrEmoteNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			server := fakeserver.New()
			defer server.Close()

			server.Resps = []string{tc.useResp}
			client := seventv.NewClient(ctx, *server.URL(t), "" /* accessToken */)

			err := client.RemoveEmote(ctx, "fake-emote-set-1", "fake-emote-1")
			if err != nil && tc.wantErr == nil {
				t.Fatalf("RemoveEmote() unexpected error: %v", err)
			}
			if err == nil && tc.wantErr != nil {
				t.Fatal("RemoveEmote() had no err, but expected one")
			}
			if err != nil && tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("RemoveEmote() expected err %v to be %v", err, tc.wantErr)
				}
			}

			const wantBody = `{"query":` +
				`"mutation ($action:ListItemAction!$emote_id:ObjectID!$emote_set_id:ObjectID!){` +
				`emoteSet(id: $emote_set_id){id,emotes(id: $emote_id, action: $action){id,name}}` +
				`}",` +
				`"variables":{"action":"REMOVE","emote_id":"fake-emote-1","emote_set_id":"fake-emote-set-1"}` +
				"}\n"

			gotBody, err := io.ReadAll(server.Reqs[0].Body)
			if err != nil {
				t.Fatalf("RemoveEmote() failed to read req body: %v", err)
			}
			if diff := cmp.Diff(wantBody, string(gotBody)); diff != "" {
				t.Errorf("RemoveEmote() req body diff (-want +got):\n%s", diff)
			}
		})
	}
}

package admin_test

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/airforce270/airbot/apiclients/bible"
	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/apiclients/kick"
	"github.com/airforce270/airbot/apiclients/seventv"
	"github.com/airforce270/airbot/apiclients/twitchtest"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/cache/cachetest"
	"github.com/airforce270/airbot/commands"
	"github.com/airforce270/airbot/commands/commandtest"
	"github.com/airforce270/airbot/config"
	"github.com/airforce270/airbot/database/databasetest"
	"github.com/airforce270/airbot/permission"
	"github.com/airforce270/airbot/platforms/twitch"
	"github.com/airforce270/airbot/testing/fakeserver"
	"github.com/pelletier/go-toml/v2"
)

func TestAdminCommands(t *testing.T) {
	t.Parallel()
	tests := []commandtest.Case{
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$botslowmode on",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "Enabled bot slowmode on Twitch",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$botslowmode off",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			Platform:  commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{enableBotSlowmode},
			Want: []*base.Message{
				{
					Text:    "Disabled bot slowmode on Twitch",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$botslowmode",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			Platform:  commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{enableBotSlowmode},
			Want: []*base.Message{
				{
					Text:    "Bot slowmode is currently enabled on Twitch",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$botslowmode on",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want:     nil,
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$echo say something",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "say something",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$echo say something",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
			},
			Platform: commandtest.TwitchPlatform,
			Want:     nil,
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$join",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			ApiResp:  twitchtest.GetChannelInformationResp,
			Want: []*base.Message{
				{
					Text:    "Successfully joined channel user1 with prefix $",
					Channel: "user2",
				},
				{
					Text:    "Successfully joined channel! (prefix: $ ) For all commands, type $commands.",
					Channel: "user1",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$join &",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			ApiResp:  twitchtest.GetChannelInformationResp,
			Want: []*base.Message{
				{
					Text:    "Successfully joined channel user1 with prefix &",
					Channel: "user2",
				},
				{
					Text:    "Successfully joined channel! (prefix: & ) For all commands, type &commands.",
					Channel: "user1",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$join",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:  commandtest.TwitchPlatform,
			ApiResp:   twitchtest.GetChannelInformationResp,
			RunBefore: []commandtest.SetupFunc{joinOtherUser1},
			Want: []*base.Message{
				{
					Text:    "Channel user1 is already joined",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$joinother user1",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want:     nil,
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$joinother user1",
					User:    "user3",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			Platform: commandtest.TwitchPlatform,
			ApiResp:  twitchtest.GetChannelInformationResp,
			Want: []*base.Message{
				{
					Text:    "Successfully joined channel user1 with prefix $",
					Channel: "user2",
				},
				{
					Text:    "Successfully joined channel! (prefix: $ ) For all commands, type $commands.",
					Channel: "user1",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$joinother user1 *",
					User:    "user3",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			Platform: commandtest.TwitchPlatform,
			ApiResp:  twitchtest.GetChannelInformationResp,
			Want: []*base.Message{
				{
					Text:    "Successfully joined channel user1 with prefix *",
					Channel: "user2",
				},
				{
					Text:    "Successfully joined channel! (prefix: * ) For all commands, type *commands.",
					Channel: "user1",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$joinother user1",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			Platform:  commandtest.TwitchPlatform,
			ApiResp:   twitchtest.GetChannelInformationResp,
			RunBefore: []commandtest.SetupFunc{joinOtherUser1},
			Want: []*base.Message{
				{
					Text:    "Channel user1 is already joined",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$joined",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			Platform:  commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{joinOtherUser1},
			Want: []*base.Message{
				{
					Text:    "Bot is currently in user1",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$joined",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want:     nil,
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$leave",
					UserID:  "user1",
					User:    "user1",
					Channel: "user1",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Admin,
			},
			Platform:  commandtest.TwitchPlatform,
			ApiResp:   twitchtest.GetChannelInformationResp,
			RunBefore: []commandtest.SetupFunc{joinOtherUser1},
			Want: []*base.Message{
				{
					Text:    "Successfully left channel.",
					Channel: "user1",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$leave",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want:     nil,
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$leaveother user1",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			Platform:  commandtest.TwitchPlatform,
			ApiResp:   twitchtest.GetChannelInformationResp,
			RunBefore: []commandtest.SetupFunc{joinOtherUser1},
			Want: []*base.Message{
				{
					Text:    "Successfully left channel user1",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$leaveother user1",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "Bot is not in channel user1",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$leaveother user1",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want:     nil,
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$reloadconfig",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			Platform:   commandtest.TwitchPlatform,
			ConfigData: "[platforms.kick]\nuser_agent = \"asdf\"",
			Want: []*base.Message{
				{
					Text:    "Reloaded config.",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$reloadconfig",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want:     nil,
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$restart",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "Restarting Airbot.",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$restart",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want:     nil,
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$setprefix &",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "Prefix set to &",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$setprefix &",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want:     nil,
		},
	}

	commandtest.Run(t, tests)
}

func TestReloadConfig(t *testing.T) {
	t.Parallel()
	server := fakeserver.New()
	defer server.Close()

	db := databasetest.New(t)
	cdb := cachetest.NewInMemory()

	platform := twitch.NewForTesting(server.URL(), db)

	const want = "something-specific"
	cfg := func() string {
		var buf strings.Builder
		enc := toml.NewEncoder(&buf)
		cfg := &config.Config{
			Platforms: config.PlatformConfig{
				Kick: config.KickConfig{
					UserAgent: want,
				},
			},
		}
		if err := enc.Encode(cfg); err != nil {
			t.Fatalf("Failed to encode %+v: %v", cfg, err)
		}
		return buf.String()
	}()

	resources := base.Resources{
		Platform: platform,
		DB:       db,
		Cache:    cdb,
		AllPlatforms: map[string]base.Platform{
			platform.Name(): platform,
		},
		NewConfigSource: func() (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(cfg)), nil
		},
		Rand: base.RandResources{
			Reader: bytes.NewBuffer([]byte{3}),
			Source: fakeExpRandSource{Value: uint64(150)},
		},
		Clients: base.APIClients{
			Bible:                         bible.NewClient(server.URL()),
			IVR:                           ivr.NewClient(server.URL()),
			Kick:                          kick.NewClient(server.URL(), "" /* ja3 */, "" /* userAgent */),
			PastebinFetchPasteURLOverride: server.URL(),
			SevenTV:                       seventv.NewClient(server.URL()),
		},
	}

	input := base.IncomingMessage{
		Message: base.Message{
			Text:    "$reloadconfig",
			UserID:  "user1",
			User:    "user1",
			Channel: "user2",
			Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
		},
		Prefix:          "$",
		PermissionLevel: permission.Owner,
		Resources:       resources,
	}

	handler := commands.NewHandlerForTest(db, cdb, resources.AllPlatforms, resources.NewConfigSource, resources.Rand, resources.Clients)

	_, err := handler.Handle(&input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := handler.Clients.Kick.UserAgent; got != want {
		t.Errorf("config value after updating = %q, want %q", got, want)
	}
}

var testConfig = config.Config{}

func joinOtherUser1(t testing.TB, r *base.Resources) {
	t.Helper()
	handler := commands.NewHandler(r.DB, r.Cache, &testConfig, nil)
	_, err := handler.Handle(&base.IncomingMessage{
		Message: base.Message{
			Text:    "$joinother user1",
			UserID:  "user1",
			User:    "user1",
			Channel: "user2",
			Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
		},
		Prefix:          "$",
		PermissionLevel: permission.Owner,
		Resources:       *r,
	})
	if err != nil {
		t.Fatalf("Failed to joinother user1: %v", err)
	}
}

func enableBotSlowmode(t testing.TB, r *base.Resources) {
	t.Helper()
	handler := commands.NewHandler(r.DB, r.Cache, &testConfig, nil)
	_, err := handler.Handle(&base.IncomingMessage{
		Message: base.Message{
			Text:    "$botslowmode on",
			UserID:  "user1",
			User:    "user1",
			Channel: "user2",
			Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
		},
		Prefix:          "$",
		PermissionLevel: permission.Owner,
		Resources:       *r,
	})
	if err != nil {
		t.Fatalf("Failed to enable bot slowmode: %v", err)
	}
}

type fakeExpRandSource struct {
	Value uint64
}

func (s fakeExpRandSource) Uint64() uint64  { return s.Value }
func (s fakeExpRandSource) Seed(val uint64) {}

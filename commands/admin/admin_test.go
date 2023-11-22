package admin_test

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/airforce270/airbot/apiclients/kick"
	"github.com/airforce270/airbot/apiclients/twitchtest"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands"
	"github.com/airforce270/airbot/commands/commandtest"
	"github.com/airforce270/airbot/config"
	"github.com/airforce270/airbot/permission"
	"github.com/pelletier/go-toml/v2"
)

func TestAdminCommands(t *testing.T) {
	config.OSReadFile = func(name string) ([]byte, error) {
		var buf bytes.Buffer
		enc := toml.NewEncoder(&buf)
		if err := enc.Encode(&config.Config{}); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	defer func() { config.OSReadFile = os.ReadFile }()
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
			Platform: commandtest.TwitchPlatform,
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
	const want = "something-specific"
	config.OSReadFile = func(name string) ([]byte, error) {
		var buf bytes.Buffer
		enc := toml.NewEncoder(&buf)
		cfg := &config.Config{
			Platforms: config.PlatformConfig{
				Kick: config.KickConfig{
					UserAgent: want,
				},
			},
		}
		if err := enc.Encode(cfg); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	defer func() { config.OSReadFile = os.ReadFile }()
	tc := commandtest.Case{
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
		Platform: commandtest.TwitchPlatform,
		Want: []*base.Message{
			{
				Text:    "Reloaded config.",
				Channel: "user2",
			},
		},
	}

	commandtest.Run(t, []commandtest.Case{tc})

	if got := *kick.UserAgent.Load(); got != want {
		t.Errorf("reloadConfig() value = %q, want %q", got, want)
	}
}

func joinOtherUser1(t testing.TB, r *base.Resources) {
	t.Helper()
	handler := commands.NewHandler(r.DB, r.Cache, r.Rand)
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
	handler := commands.NewHandler(r.DB, r.Cache, r.Rand)
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

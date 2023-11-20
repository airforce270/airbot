package commands

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/airforce270/airbot/apiclients/bible"
	"github.com/airforce270/airbot/apiclients/bible/bibletest"
	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/apiclients/ivr/ivrtest"
	"github.com/airforce270/airbot/apiclients/kick"
	"github.com/airforce270/airbot/apiclients/kick/kicktest"
	"github.com/airforce270/airbot/apiclients/pastebin"
	"github.com/airforce270/airbot/apiclients/pastebin/pastebintest"
	"github.com/airforce270/airbot/apiclients/seventv"
	"github.com/airforce270/airbot/apiclients/seventv/seventvtest"
	"github.com/airforce270/airbot/apiclients/twitchtest"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/cache"
	"github.com/airforce270/airbot/cache/cachetest"
	"github.com/airforce270/airbot/config"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/databasetest"
	"github.com/airforce270/airbot/database/models"
	"github.com/airforce270/airbot/permission"
	"github.com/airforce270/airbot/platforms/twitch"
	"github.com/airforce270/airbot/testing/fakeserver"

	"github.com/google/go-cmp/cmp"
	"github.com/nicklaw5/helix/v2"
	"github.com/pelletier/go-toml/v2"
	"gorm.io/gorm"
)

type testCase struct {
	input      base.IncomingMessage
	otherTexts []string
	apiResp    string
	apiResps   []string
	runBefore  []func() error
	runAfter   []func() error
	want       []*base.Message
	// `wantWrapped` shouldn't be set in test cases directly,
	// it's just `want` values wrapped in OutgoingMessages
	// which is set by pretest processing
	wantWrapped []*base.OutgoingMessage
}

func TestCommands(t *testing.T) {
	server := fakeserver.New()
	defer server.Close()

	config.OSReadFile = func(name string) ([]byte, error) {
		var buf bytes.Buffer
		enc := toml.NewEncoder(&buf)
		if err := enc.Encode(&config.Config{}); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	defer func() { config.OSReadFile = os.ReadFile }()

	tests := []testCase{
		// admin.go commands
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$botslowmode on",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "Enabled bot slowmode on Twitch",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$botslowmode off",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{enableBotSlowmode},
			want: []*base.Message{
				{
					Text:    "Disabled bot slowmode on Twitch",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$botslowmode",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				enableBotSlowmode,
			},
			want: []*base.Message{
				{
					Text:    "Bot slowmode is currently enabled on Twitch",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$botslowmode on",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: nil,
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$echo say something",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "say something",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
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
			want: nil,
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$join",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			apiResp: twitchtest.GetChannelInformationResp,
			want: []*base.Message{
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
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$join &",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			apiResp: twitchtest.GetChannelInformationResp,
			want: []*base.Message{
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
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$join",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			apiResp:   twitchtest.GetChannelInformationResp,
			runBefore: []func() error{joinOtherUser1},
			want: []*base.Message{
				{
					Text:    "Channel user1 is already joined",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
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
			want: nil,
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$joinother user1",
					User:    "user3",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			apiResp: twitchtest.GetChannelInformationResp,
			want: []*base.Message{
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
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$joinother user1 *",
					User:    "user3",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			apiResp: twitchtest.GetChannelInformationResp,
			want: []*base.Message{
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
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$joinother user1",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			apiResp:   twitchtest.GetChannelInformationResp,
			runBefore: []func() error{joinOtherUser1},
			want: []*base.Message{
				{
					Text:    "Channel user1 is already joined",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$joined",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{joinOtherUser1},
			want: []*base.Message{
				{
					Text:    "Bot is currently in user1",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
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
			want: nil,
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$leave",
					UserID:  "user1",
					User:    "user1",
					Channel: "user1",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Admin,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			apiResp:   twitchtest.GetChannelInformationResp,
			runBefore: []func() error{joinOtherUser1},
			want: []*base.Message{
				{
					Text:    "Successfully left channel.",
					Channel: "user1",
				},
			},
		},
		{
			input: base.IncomingMessage{
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
			want: nil,
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$leaveother user1",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			apiResp:   twitchtest.GetChannelInformationResp,
			runBefore: []func() error{joinOtherUser1},
			want: []*base.Message{
				{
					Text:    "Successfully left channel user1",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$leaveother user1",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "Bot is not in channel user1",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
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
			want: nil,
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$reloadconfig",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "Reloaded config.",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$reloadconfig",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: nil,
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$setprefix &",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "Prefix set to &",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
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
			want: nil,
		},

		// botinfo.go commands
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$bot",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{
				"$botinfo",
				"$info",
				"$about",
			},
			want: []*base.Message{
				{
					Text:    "Beep boop, this is Airbot running as fake-username in user2 with prefix $ on Twitch. Made by airforce2700, source available on GitHub ( $source )",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$help",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "For help with a command, use $help <command>. To see available commands, use $commands",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$help join",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "[ $join ] Tells the bot to join your chat.",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$help duel",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "[ $duel ] Duels another chatter. They have 30 seconds to accept or decline. User-specific cooldown: 5s",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$help pyramid",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "[ $pyramid ] Makes a pyramid in chat. Max width 25. Channel-wide cooldown: 30s",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "??prefix",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "??",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "This channel's prefix is ??",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    ";prefix",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			otherTexts: []string{
				"does this bot thingy have one of them prefixes",
				"what is a prefix",
				"forsen prefix",
				"Successfully joined channel iP0G with prefix $",
			},
			want: nil,
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$source",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "Source code for Airbot available at https://github.com/airforce270/airbot",
					Channel: "user2",
				},
			},
		},
		// stats is currently untested due to reliance on low-level syscalls

		// bulk.go commands
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$filesay https://pastebin.com/raw/B7TBjQEy",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			apiResp: pastebintest.MultiLineFetchPasteResp,
			want: []*base.Message{
				{
					Text:    "line1",
					Channel: "user2",
				},
				{
					Text:    "line2",
					Channel: "user2",
				},
				{
					Text:    "line3",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$filesay https://pastebin.com/raw/B7TBjQEy",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: pastebintest.MultiLineFetchPasteResp,
			want:    nil,
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$filesay",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			apiResp: pastebintest.MultiLineFetchPasteResp,
			want: []*base.Message{
				{
					Text:    "Usage: $filesay <pastebin raw URL>",
					Channel: "user2",
				},
			},
		},

		// echo.go commands
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$commands",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "Commands available here: https://github.com/airforce270/airbot/blob/main/docs/commands.md",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$gn",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "FeelsOkayMan <3 gn user1",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$spam 3 yo",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "yo",
					Channel: "user2",
				},
				{
					Text:    "yo",
					Channel: "user2",
				},
				{
					Text:    "yo",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$spam 3 yo",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: nil,
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$pyramid 5 yo",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "yo",
					Channel: "user2",
				},
				{
					Text:    "yo yo",
					Channel: "user2",
				},
				{
					Text:    "yo yo yo",
					Channel: "user2",
				},
				{
					Text:    "yo yo yo yo",
					Channel: "user2",
				},
				{
					Text:    "yo yo yo yo yo",
					Channel: "user2",
				},
				{
					Text:    "yo yo yo yo",
					Channel: "user2",
				},
				{
					Text:    "yo yo yo",
					Channel: "user2",
				},
				{
					Text:    "yo yo",
					Channel: "user2",
				},
				{
					Text:    "yo",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$pyramid 1000 yo",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "Max pyramid width is 25",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$pyramid 5 yo",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: nil,
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$trihard",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$TriHard"},
			want: []*base.Message{
				{
					Text:    "TriHard 7",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$tuck",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "Usage: $tuck <user>",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$tuck someone",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "Bedge user1 tucks someone into bed.",
					Channel: "user2",
				},
			},
		},

		// fun.go commands
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$bibleverse Philippians 4:8",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$bv Philippians 4:8"},
			apiResp:    bibletest.LookupVerseSingleVerse1Resp,
			want: []*base.Message{
				{
					Text:    "[Philippians 4:8]: Finally, brothers, whatever things are true, whatever things are honorable, whatever things are just, whatever things are pure, whatever things are lovely, whatever things are of good report; if there is any virtue, and if there is any praise, think about these things.\n",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$bibleverse John 3:16",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$bv John 3:16"},
			apiResp:    bibletest.LookupVerseSingleVerse2Resp,
			want: []*base.Message{
				{
					Text:    "[John 3:16]: \nFor God so loved the world, that he gave his one and only Son, that whoever believes in him should not perish, but have eternal life.\n\n",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$bibleverse",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$bv"},
			want: []*base.Message{
				{
					Text:    "Usage: $bibleverse <book> <chapter:verse>",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$cock",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "user1's cock is 3 inches long",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$cock someone",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "someone's cock is 3 inches long",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$iq",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "user1's IQ is 100",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$iq someone",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "someone's IQ is 100",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$ship person1 person2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				func() error {
					base.RandReader = bytes.NewBuffer([]byte{95})
					return nil
				},
			},
			want: []*base.Message{
				{
					Text:    "person1 and person2 have a 95% compatibility, invite me to the wedding please 😍",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$ship person1 person2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				func() error {
					base.RandReader = bytes.NewBuffer([]byte{85})
					return nil
				},
			},
			want: []*base.Message{
				{
					Text:    "person1 and person2 have a 85% compatibility, oh 😳",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$ship person1 person2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				func() error {
					base.RandReader = bytes.NewBuffer([]byte{70})
					return nil
				},
			},
			want: []*base.Message{
				{
					Text:    "person1 and person2 have a 70% compatibility, worth a shot ;)",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$ship person1 person2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				func() error {
					base.RandReader = bytes.NewBuffer([]byte{50})
					return nil
				},
			},
			want: []*base.Message{
				{
					Text:    "person1 and person2 have a 50% compatibility, it's a toss-up :/",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$ship person1 person2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				func() error {
					base.RandReader = bytes.NewBuffer([]byte{30})
					return nil
				},
			},
			want: []*base.Message{
				{
					Text:    "person1 and person2 have a 30% compatibility, not sure about this one... :(",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$ship person1 person2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				func() error {
					base.RandReader = bytes.NewBuffer([]byte{5})
					return nil
				},
			},
			want: []*base.Message{
				{
					Text:    "person1 and person2 have a 5% compatibility, don't even think about it DansGame",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$ship person1",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "Usage: $ship <first-person> <second-person>",
					Channel: "user2",
				},
			},
		},

		// gamba.go commands
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$accept",
					UserID:  "user2",
					User:    "user2",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				setRandValueTo1,
				add50PointsToUser1,
				add50PointsToUser2,
				startDuel,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "user1 won the duel with user2 and wins 25 points!",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$accept",
					UserID:  "user2",
					User:    "user2",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				setRandValueTo0,
				add50PointsToUser1,
				add50PointsToUser2,
				startDuel,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "user2 won the duel with user1 and wins 25 points!",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$accept",
					UserID:  "user2",
					User:    "user2",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				setRandValueTo1,
				add50PointsToUser1,
				add50PointsToUser2,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "There are no duels pending against you.",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$decline",
					UserID:  "user2",
					User:    "user2",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				setRandValueTo1,
				add50PointsToUser1,
				add50PointsToUser2,
				startDuel,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "Declined duel.",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$decline",
					UserID:  "user2",
					User:    "user2",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				setRandValueTo1,
				add50PointsToUser1,
				add50PointsToUser2,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "There are no duels pending against you.",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2 25",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				add50PointsToUser1,
				add50PointsToUser2,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "@user2, user1 has started a duel for 25 points! Type $accept or $decline in the next 30 seconds!",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2 25",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				add50PointsToUser2,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "You don't have enough points for that duel (you have 0 points)",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2 25",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				add50PointsToUser1,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "user2 don't have enough points for that duel (they have 0 points)",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2 25",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				add50PointsToUser1,
				add50PointsToUser2,
				startDuel,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "You already have a duel pending.",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2 25",
					UserID:  "user3",
					User:    "user3",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				add50PointsToUser1,
				add50PointsToUser2,
				add50PointsToUser3,
				startDuel,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "That chatter already has a duel pending.",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "Usage: $duel <user> <amount>",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "Usage: $duel <user> <amount>",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user1 10",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "You can't duel yourself Pepega",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2 0",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "You must duel at least 1 point.",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2 xx",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "Usage: $duel <user> <amount>",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$givepoints user2 10",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				add50PointsToUser1,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "user1 gave 10 points to user2 FeelsOkayMan <3",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$givepoints user2 100",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				add50PointsToUser1,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "You can't give more points than you have (you have 50 points)",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$givepoints user2 0",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				add50PointsToUser1,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "You must give at least 1 point.",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$givepoints user2 xx",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				add50PointsToUser1,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "Usage: $givepoints <user> <amount>",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$givepoints user2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "Usage: $givepoints <user> <amount>",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$givepoints",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "Usage: $givepoints <user> <amount>",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$points",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{
				"$points user1",
				"$p",
				"$p user1",
			},
			runBefore: []func() error{
				add50PointsToUser1,
			},
			want: []*base.Message{
				{
					Text:    "GAMBA user1 has 50 points",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$points user1",
					UserID:  "user2",
					User:    "user2",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$p user1"},
			runBefore: []func() error{
				add50PointsToUser1,
			},
			want: []*base.Message{
				{
					Text:    "GAMBA user1 has 50 points",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$points rando",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$p rando"},
			want: []*base.Message{
				{
					Text:    "rando has never been seen by fake-username",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$roulette 10",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{
				"$r 10",
				"$roulette 20%",
				"$r 20%",
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				setRandValueTo1,
				add50PointsToUser1,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "GAMBA user1 won 10 points in roulette and now has 60 points!",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$roulette 10",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{
				"$r 10",
				"$roulette 20%",
				"$r 20%",
			},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				setRandValueTo0,
				add50PointsToUser1,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "GAMBA user1 lost 10 points in roulette and now has 40 points!",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$roulette all",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$r all"},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				setRandValueTo1,
				add50PointsToUser1,
			},
			runAfter: []func() error{
				waitForTransactionsToSettle,
			},
			want: []*base.Message{
				{
					Text:    "GAMBA user1 won 50 points in roulette and now has 100 points!",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$roulette 60",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$r 60"},
			runBefore: []func() error{
				deleteAllGambaTransactions,
				setRandValueTo0,
				add50PointsToUser1,
			},
			want: []*base.Message{
				{
					Text:    "user1: You don't have enough points for that (current: 50)",
					Channel: "user2",
				},
			},
		},

		// kick.go commands
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$kickislive brucedropemoff",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$kislive brucedropemoff"},
			apiResp:    kicktest.LargeLiveGetChannelResp,
			want: []*base.Message{
				{
					Text:    "brucedropemoff is currently live on Kick, streaming Just Chatting to 15274 viewers.",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$kickislive xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$kislive xqc"},
			apiResp:    kicktest.LargeOfflineGetChannelResp,
			want: []*base.Message{
				{
					Text:    "xqc is not currently live on Kick.",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$kicktitle brucedropemoff",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$ktitle brucedropemoff"},
			apiResp:    kicktest.LargeLiveGetChannelResp,
			want: []*base.Message{
				{
					Text:    "brucedropemoff's title on Kick: DEO COOKOFF MAY THE BEST DISH WIN! 🗣️ #DEO4L",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$kicktitle xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$ktitle xqc"},
			apiResp:    kicktest.LargeOfflineGetChannelResp,
			want: []*base.Message{
				{
					Text:    "Currently Kick only returns the title for live channels, and xqc is not currently live.",
					Channel: "user2",
				},
			},
		},

		// moderation.go commands
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$vanish",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			runAfter: []func() error{
				waitForMessagesToSend,
			},
			want: nil,
		},

		// seventv.go commands
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv emotecount",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResps: []string{
				twitchtest.GetUsersResp,
				seventvtest.FetchUserConnectionByTwitchUserIdSmallNonSubResp,
			},
			want: []*base.Message{
				{
					Text:    "user1 has 3 emotes on 7TV",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$7tv emotecount airforce2700",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResps: []string{
				twitchtest.GetUsersResp,
				seventvtest.FetchUserConnectionByTwitchUserIdSmallNonSubResp,
			},
			want: []*base.Message{
				{
					Text:    "airforce2700 has 3 emotes on 7TV",
					Channel: "user2",
				},
			},
		},

		// twitch.go commands
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$banreason",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$br"},
			apiResp:    ivrtest.TwitchUsersBannedResp,
			want: []*base.Message{
				{
					Text:    "Usage: $banreason <user>",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$banreason banneduser",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$br banneduser"},
			apiResp:    ivrtest.TwitchUsersBannedResp,
			want: []*base.Message{
				{
					Text:    "SeaGrade's ban reason: TOS_INDEFINITE",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$banreason nonbanneduser",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$br nonbanneduser"},
			apiResp:    ivrtest.TwitchUsersNotStreamingResp,
			want: []*base.Message{
				{
					Text:    "xQc is not banned.",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$currentgame",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: twitchtest.GetChannelInformationResp,
			want: []*base.Message{
				{
					Text:    "user1 is currently playing Science&Technology",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$founders",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.FoundersNormalResp,
			want: []*base.Message{
				{
					Text:    "user2's founders are: FishyyKingyy, eljulidi1337, SamMist, Leochansz, lexieuzumaki7, ContraVz, rott______, DankJuicer, kronikZ____, blemplob",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$founders hasfounders",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.FoundersNormalResp,
			want: []*base.Message{
				{
					Text:    "hasfounders's founders are: FishyyKingyy, eljulidi1337, SamMist, Leochansz, lexieuzumaki7, ContraVz, rott______, DankJuicer, kronikZ____, blemplob",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$founders nofounders",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.FoundersNoneResp,
			want: []*base.Message{
				{
					Text:    "nofounders has no founders",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$founders nofounders404",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.FoundersNone404Resp,
			want: []*base.Message{
				{
					Text:    "nofounders404 has no founders",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$subage macroblank1 xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.SubAgeCurrentPaidTier3Resp,
			want: []*base.Message{
				{
					Text:    "Macroblank1 is currently subscribed to xQc with a Tier 3 paid subscription (3 days remaining) and is on a 17 month streak",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$sa macroblank1 xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.SubAgeCurrentPaidTier3Resp,
			want: []*base.Message{
				{
					Text:    "Macroblank1 is currently subscribed to xQc with a Tier 3 paid subscription (3 days remaining) and is on a 17 month streak",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$sublength macroblank1 xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.SubAgeCurrentPaidTier3Resp,
			want: []*base.Message{
				{
					Text:    "Macroblank1 is currently subscribed to xQc with a Tier 3 paid subscription (3 days remaining) and is on a 17 month streak",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$subage ellagarten xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.SubAgeCurrentGiftTier1Resp,
			want: []*base.Message{
				{
					Text:    "ellagarten is currently subscribed to xQc with a Tier 1 gifted subscription (14 days remaining) and is on a 4 month streak (total: 17 months)",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$subage airforce2700 xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.SubAgeCurrentPrimeResp,
			want: []*base.Message{
				{
					Text:    "airforce2700 is currently subscribed to xQc with a Prime subscription (1 day remaining) and is on a 22 month streak",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$subage @airforce2700 @elis",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.SubAgePreviousSubResp,
			want: []*base.Message{
				{
					Text:    "airforce2700 is not currently subscribed to elis, but was previously subscribed for 4 months",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$subage airforce2700 hasanabi",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.SubAgeNeverSubbedResp,
			want: []*base.Message{
				{
					Text:    "airforce2700 is not subscribed to HasanAbi and has not been previously subscribed",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$subage airforce2700 channelthatdoesntexist",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.SubAge404ChannelResp,
			want: []*base.Message{
				{
					Text:    "Channel channelthatdoesntexist was not found",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$subage userthatdoesntexist xqc",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.SubAge404UserResp,
			want: []*base.Message{
				{
					Text:    "User userthatdoesntexist was not found",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$logs xqc forsen",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "forsen's logs in xqc's chat: https://logs.ivr.fi/?channel=xqc&username=forsen",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$logs",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			want: []*base.Message{
				{
					Text:    "Usage: $logs <channel> <user>",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$mods",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*base.Message{
				{
					Text:    "user2's mods are: StreamElements, Fossabot, spintto, HNoAce",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$mods otherchannel",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*base.Message{
				{
					Text:    "otherchannel's mods are: StreamElements, Fossabot, spintto, HNoAce",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$mods nomods",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.ModsAndVIPsNoneResp,
			want: []*base.Message{
				{
					Text:    "nomods has no mods",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$title",
					UserID:  "user2",
					User:    "user2",
					Channel: "user1",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$title user1"},
			apiResp:    twitchtest.GetChannelInformationResp,
			want: []*base.Message{
				{
					Text:    "user1's title: TwitchDevMonthlyUpdate//May6,2021",
					Channel: "user1",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$verifiedbot",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{
				"$verifiedbot otherchannel",
				"$vb",
				"$vb otherchannel",
			},
			apiResp: ivrtest.TwitchUsersVerifiedBotResp,
			want: []*base.Message{
				{
					Text:    "This command is currently offline due to changes on Twitch's end :(",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$verifiedbot notverified",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{"$vb notverified"},
			apiResp:    ivrtest.TwitchUsersNotVerifiedBotResp,
			want: []*base.Message{
				{
					Text:    "This command is currently offline due to changes on Twitch's end :(",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$verifiedbotquiet",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{
				"$verifiedbotquiet otherchannel",
				"$verifiedbotq",
				"$verifiedbotq otherchannel",
				"$vbquiet",
				"$vbquiet otherchannel",
				"$vbq",
				"$vbq otherchannel",
			},
			apiResp: ivrtest.TwitchUsersVerifiedBotResp,
			want: []*base.Message{
				{
					Text:    "This command is currently offline due to changes on Twitch's end :(",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$verifiedbotquiet notverified",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			otherTexts: []string{
				"$verifiedbotq notverified",
				"$vbquiet notverified",
				"$vbq notverified",
			},
			apiResp: ivrtest.TwitchUsersNotVerifiedBotResp,
			want: []*base.Message{
				{
					Text:    "This command is currently offline due to changes on Twitch's end :(",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$vips",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*base.Message{
				{
					Text:    "user2's VIPs are: bakonsword, alyjiahT_T, AVBest, Zaintew_, captkayy, seagrad, Dafkeee",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$vips otherchannel",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*base.Message{
				{
					Text:    "otherchannel's VIPs are: bakonsword, alyjiahT_T, AVBest, Zaintew_, captkayy, seagrad, Dafkeee",
					Channel: "user2",
				},
			},
		},
		{
			input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$vips novips",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDBConn()),
			},
			apiResp: ivrtest.ModsAndVIPsNoneResp,
			want: []*base.Message{
				{
					Text:    "novips has no VIPs",
					Channel: "user2",
				},
			},
		},
	}

	for _, unbuiltTC := range tests {
		for _, tc := range buildTestCases(t, unbuiltTC) {
			t.Run(fmt.Sprintf("[%s] %s", tc.input.PermissionLevel.Name(), tc.input.Message.Text), func(t *testing.T) {
				server.Resp = tc.apiResp
				server.Resps = tc.apiResps
				db := databasetest.NewFakeDB(t)
				database.SetInstance(db)
				setFakes(server.URL(), db)
				for i, f := range tc.runBefore {
					if err := f(); err != nil {
						t.Fatalf("runBefore[%d] func failed: %v", i, err)
					}
				}

				handler := Handler{db: db}
				got, err := handler.Handle(&tc.input)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				for i, f := range tc.runAfter {
					if err := f(); err != nil {
						t.Fatalf("runAfter[%d] func failed: %v", i, err)
					}
				}

				if diff := cmp.Diff(tc.wantWrapped, got); diff != "" {
					t.Errorf("Handle() diff (-want +got):\n%s", diff)
				}
				resetFakes()
				server.Reset()
			})
		}
	}
}

func buildTestCases(t *testing.T, tc testCase) []testCase {
	for _, want := range tc.want {
		tc.wantWrapped = append(tc.wantWrapped, &base.OutgoingMessage{Message: *want})
	}
	tcs := []testCase{tc}
	for _, otherText := range tc.otherTexts {
		tcCopy := tc
		tcCopy.input.Message.Text = otherText
		tcs = append(tcs, tcCopy)
	}
	return tcs
}

var (
	savedBibleURL = bible.BaseURL
	savedIVRURL   = ivr.BaseURL
	savedKickURL  = kick.BaseURL
	saved7TVURL   = seventv.BaseURL
)

type fakeExpRandSource struct {
	Value uint64
}

func (s fakeExpRandSource) Uint64() uint64  { return s.Value }
func (s fakeExpRandSource) Seed(val uint64) {}

func setFakes(url string, db *gorm.DB) {
	base.RandReader = bytes.NewBuffer([]byte{3})
	base.RandSource = fakeExpRandSource{Value: uint64(150)}
	bible.BaseURL = url
	cache.SetInstance(cachetest.NewInMemory())
	ivr.BaseURL = url
	kick.BaseURL = url
	pastebin.FetchPasteURLOverride = url
	seventv.BaseURL = url
	twitch.Conn = twitch.NewForTesting(url, db)
}

func resetFakes() {
	base.RandReader = rand.Reader
	base.RandSource = nil
	bible.BaseURL = savedBibleURL
	cache.SetInstance(nil)
	ivr.BaseURL = savedIVRURL
	kick.BaseURL = savedKickURL
	pastebin.FetchPasteURLOverride = ""
	seventv.BaseURL = saved7TVURL
	twitch.Conn = twitch.NewForTesting(helix.DefaultAPIBaseURL, nil)
}

func joinOtherUser1() error {
	db := databasetest.NewFakeDBConn()
	handler := Handler{db: db}
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
		Platform:        twitch.NewForTesting("forsen", db),
	})
	if err != nil {
		return fmt.Errorf("failed to joinother user1: %w", err)
	}
	return nil
}

func enableBotSlowmode() error {
	db := databasetest.NewFakeDBConn()
	handler := Handler{db: db}
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
		Platform:        twitch.NewForTesting("forsen", db),
	})
	if err != nil {
		return fmt.Errorf("failed to enable bot slowmode: %w", err)
	}
	return nil
}

func setRandValueTo0() error {
	base.RandReader = bytes.NewBuffer([]byte{0})
	return nil
}

func setRandValueTo1() error {
	base.RandReader = bytes.NewBuffer([]byte{1})
	return nil
}

func waitForMessagesToSend() error {
	time.Sleep(20 * time.Millisecond)
	return nil
}

func waitForTransactionsToSettle() error {
	time.Sleep(20 * time.Millisecond)
	return nil
}

func deleteAllGambaTransactions() error {
	db := databasetest.NewFakeDBConn()
	err := db.Where("1=1").Delete(&models.GambaTransaction{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete all gamba txns: %w", err)
	}
	return nil
}

func startDuel() error {
	db := databasetest.NewFakeDBConn()
	var user1, user2 models.User
	err := db.First(&user1, models.User{
		TwitchID:   "user1",
		TwitchName: "user1",
	}).Error
	if err != nil {
		return fmt.Errorf("failed to find user1: %w", err)
	}
	err = db.First(&user2, models.User{
		TwitchID:   "user2",
		TwitchName: "user2",
	}).Error
	if err != nil {
		return fmt.Errorf("failed to find user2: %w", err)
	}
	err = db.Create(&models.Duel{
		UserID:   user1.ID,
		User:     user1,
		TargetID: user2.ID,
		Target:   user2,
		Amount:   25,
		Pending:  true,
		Accepted: false,
	}).Error
	if err != nil {
		return fmt.Errorf("failed to create duel: %w", err)
	}
	return nil
}

func add50PointsToUser1() error {
	db := databasetest.NewFakeDBConn()
	var user models.User
	err := db.First(&user, models.User{
		TwitchID:   "user1",
		TwitchName: "user1",
	}).Error
	if err != nil {
		return fmt.Errorf("failed to find user1: %w", err)
	}
	return add50PointsToUser(user, db)
}

func add50PointsToUser2() error {
	db := databasetest.NewFakeDBConn()
	var user models.User
	err := db.First(&user, models.User{
		TwitchID:   "user2",
		TwitchName: "user2",
	}).Error
	if err != nil {
		return fmt.Errorf("failed to find/create user2: %w", err)
	}
	return add50PointsToUser(user, db)
}

func add50PointsToUser3() error {
	db := databasetest.NewFakeDBConn()
	var user models.User
	err := db.First(&user, models.User{
		TwitchID:   "user3",
		TwitchName: "user3",
	}).Error
	if err != nil {
		return fmt.Errorf("failed to find/create user3: %w", err)
	}
	return add50PointsToUser(user, db)
}

func add50PointsToUser(user models.User, db *gorm.DB) error {
	txn := models.GambaTransaction{
		Game:  "FAKE - TEST",
		User:  user,
		Delta: 50,
	}
	if err := db.Create(&txn).Error; err != nil {
		return fmt.Errorf("failed to insert gamba transaction: %w", err)
	}
	return nil
}

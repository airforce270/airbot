package commands

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/airforce270/airbot/apiclients/bibletest"
	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/apiclients/ivrtest"
	"github.com/airforce270/airbot/apiclients/pastebin"
	"github.com/airforce270/airbot/apiclients/pastebintest"
	"github.com/airforce270/airbot/apiclients/twitchtest"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/config"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/models"
	"github.com/airforce270/airbot/permission"
	"github.com/airforce270/airbot/platforms/twitch"
	"github.com/airforce270/airbot/testing/databasetest"
	"github.com/airforce270/airbot/testing/fakeserver"

	"github.com/google/go-cmp/cmp"
	"github.com/jinzhu/copier"
	"github.com/nicklaw5/helix/v2"
	"gorm.io/gorm"
)

type testCase struct {
	input     *base.IncomingMessage
	apiResp   string
	runBefore []func() error
	want      []*base.Message
}

func TestCommands(t *testing.T) {
	server := fakeserver.New()
	defer server.Close()

	config.OSReadFile = func(name string) ([]byte, error) {
		return []byte("blahblah"), nil
	}

	tests := flatten(
		// admin.go commands
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$join",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDB()),
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
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$join",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDB()),
			},
			apiResp:   twitchtest.GetChannelInformationResp,
			runBefore: []func() error{joinOtherUser1},
			want: []*base.Message{
				{
					Text:    "Channel user1 is already joined",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
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
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$joinother user1",
					User:    "user3",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDB()),
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
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$joinother user1",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDB()),
			},
			apiResp:   twitchtest.GetChannelInformationResp,
			runBefore: []func() error{joinOtherUser1},
			want: []*base.Message{
				{
					Text:    "Channel user1 is already joined",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$joined",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDB()),
			},
			runBefore: []func() error{joinOtherUser1},
			want: []*base.Message{
				{
					Text:    "Bot is currently in user1",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
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
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$leave",
					UserID:  "user1",
					User:    "user1",
					Channel: "user1",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Admin,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDB()),
			},
			apiResp:   twitchtest.GetChannelInformationResp,
			runBefore: []func() error{joinOtherUser1},
			want: []*base.Message{
				{
					Text:    "Successfully left channel.",
					Channel: "user1",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
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
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$leaveother user1",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDB()),
			},
			apiResp:   twitchtest.GetChannelInformationResp,
			runBefore: []func() error{joinOtherUser1},
			want: []*base.Message{
				{
					Text:    "Successfully left channel user1",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$leaveother user1",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDB()),
			},
			want: []*base.Message{
				{
					Text:    "Bot is not in channel user1",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
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
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$setprefix &",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDB()),
			},
			want: []*base.Message{
				{
					Text:    "Prefix set to &",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
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
		}),

		// botinfo.go commands
		testCasesWithSameOutput([]string{
			"$bot",
			"$botinfo",
			"$info",
			"$about",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting("forsen", databasetest.NewFakeDB()),
			},
			want: []*base.Message{
				{
					Text:    "Beep boop, this is Airbot running as fake-username in user2 with prefix $ on Twitch. Made by airforce2700, source available on GitHub ( $source )",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$help",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: []*base.Message{
				{
					Text:    "For help with a command, use $help <command>. To see available commands, use $commands",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$help join",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: []*base.Message{
				{
					Text:    "[ $join ] Tells the bot to join your chat.",
					Channel: "user2",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			"??prefix",
			"prefix",
			"wats the prefix",
			"wats the prefix?",
			"whats the prefix",
			"what's the prefix",
			"whats airbot's prefix",
			"whats af2bot's prefix",
			"whats the bots prefix",
			"whats the bot's prefix",
			"what's the bots prefix",
			"what's the bot's prefix",
			"what's the bot's prefix?",
			"what is the bots prefix",
			"what is the bot's prefix",
			"yo what is the bot's prefix bro",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "??",
				PermissionLevel: permission.Normal,
			},
			want: []*base.Message{
				{
					Text:    "This channel's prefix is ??",
					Channel: "user2",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			";prefix",
			"does this bot thingy have one of them prefixes",
			"what is a prefix",
			"forsen prefix",
			"Successfully joined channel iP0G with prefix $",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: nil,
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$source",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: []*base.Message{
				{
					Text:    "Source code for airbot available at https://github.com/airforce270/airbot",
					Channel: "user2",
				},
			},
		}),
		// stats is currently untested due to reliance on low-level syscalls

		// bulk.go commands
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$filesay https://pastebin.com/raw/B7TBjQEy",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
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
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
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
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$filesay",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
			},
			apiResp: pastebintest.MultiLineFetchPasteResp,
			want: []*base.Message{
				{
					Text:    "usage: $filesay <pastebin raw url>",
					Channel: "user2",
				},
			},
		}),

		// echo.go commands
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$commands",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: []*base.Message{
				{
					Text:    "Commands available here: https://github.com/airforce270/airbot/blob/main/docs/commands.md",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$gn",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: []*base.Message{
				{
					Text:    "FeelsOkayMan <3 gn user1",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$spam 3 yo",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
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
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
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
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$pyramid 5 yo",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
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
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$pyramid 1000 yo",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Mod,
			},
			want: []*base.Message{
				{
					Text:    "Max pyramid width is 25",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
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
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$TriHard",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: []*base.Message{
				{
					Text:    "TriHard 7",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$tuck",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: nil,
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$tuck someone",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: []*base.Message{
				{
					Text:    "Bedge user1 tucks someone into bed.",
					Channel: "user2",
				},
			},
		}),

		// fun.go commands
		testCasesWithSameOutput([]string{
			"$bibleverse Philippians 4:8",
			"$bv Philippians 4:8",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: bibletest.LookupVerseSingleVerse1Resp,
			want: []*base.Message{
				{
					Text:    "[Philippians 4:8]: Finally, brothers, whatever things are true, whatever things are honorable, whatever things are just, whatever things are pure, whatever things are lovely, whatever things are of good report; if there is any virtue, and if there is any praise, think about these things.\n",
					Channel: "user2",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			"$bibleverse John 3:16",
			"$bv John 3:16",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: bibletest.LookupVerseSingleVerse2Resp,
			want: []*base.Message{
				{
					Text:    "[John 3:16]: \nFor God so loved the world, that he gave his one and only Son, that whoever believes in him should not perish, but have eternal life.\n\n",
					Channel: "user2",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			"$bibleverse",
			"$bv",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: nil,
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$cock",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDB()),
			},
			want: []*base.Message{
				{
					Text:    "user1's cock is 3 inches long",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$cock someone",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDB()),
			},
			want: []*base.Message{
				{
					Text:    "someone's cock is 3 inches long",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$iq",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDB()),
			},
			want: []*base.Message{
				{
					Text:    "user1's IQ is 100",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$iq someone",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDB()),
			},
			want: []*base.Message{
				{
					Text:    "someone's IQ is 100",
					Channel: "user2",
				},
			},
		}),

		// gamba.go commands
		testCasesWithSameOutput([]string{
			"$points",
			"$points user1",
			"$p",
			"$p user1",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDB()),
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
		}),
		testCasesWithSameOutput([]string{
			"$points user1",
			"$p user1",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user2",
					User:    "user2",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDB()),
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
		}),
		testCasesWithSameOutput([]string{
			"$points rando",
			"$p rando",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDB()),
			},
			want: []*base.Message{
				{
					Text:    "rando has never been seen by fake-username",
					Channel: "user2",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			"$roulette 10",
			"$r 10",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDB()),
			},
			runBefore: []func() error{
				setRandValueTo1,
				add50PointsToUser1,
			},
			want: []*base.Message{
				{
					Text:    "GAMBA user1 won 10 points in roulette and now has 60 points!",
					Channel: "user2",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			"$roulette 10",
			"$r 10",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDB()),
			},
			runBefore: []func() error{
				setRandValueTo0,
				add50PointsToUser1,
			},
			want: []*base.Message{
				{
					Text:    "GAMBA user1 lost 10 points in roulette and now has 40 points!",
					Channel: "user2",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			"$roulette 60",
			"$r 60",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDB()),
			},
			runBefore: []func() error{
				setRandValueTo0,
				add50PointsToUser1,
			},
			want: []*base.Message{
				{
					Text:    "user1: You don't have enough points for that (current: 50)",
					Channel: "user2",
				},
			},
		}),

		// moderation.go commands
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$vanish",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Platform:        twitch.NewForTesting(server.URL(), databasetest.NewFakeDB()),
			},
			want: []*base.Message{
				{
					Text:    "/timeout user1 1",
					Channel: "user2",
				},
			},
		}),

		// twitch.go commands
		testCasesWithSameOutput([]string{
			"$br",
			"$banreason",
			"$banreason banneduser",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.TwitchUsersBannedResp,
			want: []*base.Message{
				{
					Text:    "SeaGrade's ban reason: TOS_INDEFINITE",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$banreason nonbanneduser",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.TwitchUsersNotStreamingResp,
			want: []*base.Message{
				{
					Text:    "xQc is not banned.",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$currentgame",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: twitchtest.GetChannelInformationResp,
			want: []*base.Message{
				{
					Text:    "user1 is currenly playing Science&Technology",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$founders",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.FoundersNormalResp,
			want: []*base.Message{
				{
					Text:    "user1's founders are: FishyyKingyy, eljulidi1337, SamMist, Leochansz, lexieuzumaki7, ContraVz, rott______, DankJuicer, kronikZ____, blemplob",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$founders hasfounders",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.FoundersNormalResp,
			want: []*base.Message{
				{
					Text:    "hasfounders's founders are: FishyyKingyy, eljulidi1337, SamMist, Leochansz, lexieuzumaki7, ContraVz, rott______, DankJuicer, kronikZ____, blemplob",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$founders nofounders",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.FoundersNoneResp,
			want: []*base.Message{
				{
					Text:    "nofounders has no founders",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$founders nofounders404",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.FoundersNone404Resp,
			want: []*base.Message{
				{
					Text:    "nofounders404 has no founders",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$logs xqc forsen",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: []*base.Message{
				{
					Text:    "forsen's logs in xqc's chat: https://logs.ivr.fi/?channel=xqc&username=forsen",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$logs",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: []*base.Message{
				{
					Text:    "Usage: $logs <channel> <user>",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$mods",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*base.Message{
				{
					Text:    "user1's mods are: StreamElements, Fossabot, spintto, HNoAce",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$mods otherchannel",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*base.Message{
				{
					Text:    "otherchannel's mods are: StreamElements, Fossabot, spintto, HNoAce",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$mods nomods",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.ModsAndVIPsNoneResp,
			want: []*base.Message{
				{
					Text:    "nomods has no mods",
					Channel: "user2",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			"$title",
			"$title otherchannel",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: twitchtest.GetChannelInformationResp,
			want: []*base.Message{
				{
					Text:    "user1's title: TwitchDevMonthlyUpdate//May6,2021",
					Channel: "user2",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			"$verifiedbot",
			"$verifiedbot otherchannel",
			"$vb",
			"$vb otherchannel",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.TwitchUsersVerifiedBotResp,
			want: []*base.Message{
				{
					Text:    "iP0G is a verified bot. ✅",
					Channel: "user2",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			"$verifiedbot notverified",
			"$vb notverified",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.TwitchUsersNotVerifiedBotResp,
			want: []*base.Message{
				{
					Text:    "xQc is not a verified bot. ❌",
					Channel: "user2",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			"$verifiedbotquiet",
			"$verifiedbotquiet otherchannel",
			"$verifiedbotq",
			"$verifiedbotq otherchannel",
			"$vbquiet",
			"$vbquiet otherchannel",
			"$vbq",
			"$vbq otherchannel",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.TwitchUsersVerifiedBotResp,
			want: []*base.Message{
				{
					Text:    "✅",
					Channel: "user2",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			"$verifiedbotquiet notverified",
			"$verifiedbotq notverified",
			"$vbquiet notverified",
			"$vbq notverified",
		}, testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.TwitchUsersNotVerifiedBotResp,
			want: []*base.Message{
				{
					Text:    "❌",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$vips",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*base.Message{
				{
					Text:    "user1's VIPs are: bakonsword, alyjiahT_T, AVBest, Zaintew_, captkayy, seagrad, Dafkeee",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$vips otherchannel",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*base.Message{
				{
					Text:    "otherchannel's VIPs are: bakonsword, alyjiahT_T, AVBest, Zaintew_, captkayy, seagrad, Dafkeee",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "$vips novips",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.ModsAndVIPsNoneResp,
			want: []*base.Message{
				{
					Text:    "novips has no VIPs",
					Channel: "user2",
				},
			},
		}),
	)

	for _, tc := range tests {
		t.Run(fmt.Sprintf("[%s] %s", tc.input.PermissionLevel.Name(), tc.input.Message.Text), func(t *testing.T) {
			server.Resp = tc.apiResp
			db := databasetest.NewFakeDB()
			database.Instance = db
			setFakes(server.URL(), db)
			for i, f := range tc.runBefore {
				if err := f(); err != nil {
					t.Fatalf("runBefore[%d] func failed: %v", i, err)
				}
			}

			handler := Handler{db: db, nonPrefixCommandsEnabled: true}
			got, err := handler.Handle(tc.input)
			if err != nil {
				fmt.Printf("unexpected error: %v\n", err)
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Handle() diff (-want +got):\n%s", diff)
			}
			resetFakes()
			server.Reset()
		})
	}

	config.OSReadFile = os.ReadFile
}

func TestCommands_EnableNonPrefixCommands(t *testing.T) {
	tests := []struct {
		input                   *base.IncomingMessage
		enableNonPrefixCommands bool
		want                    []*base.Message
	}{
		{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "whats the bots prefix",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "??",
				PermissionLevel: permission.Normal,
			},
			enableNonPrefixCommands: true,
			want: []*base.Message{
				{
					Text:    "This channel's prefix is ??",
					Channel: "user2",
				},
			},
		},
		{
			input: &base.IncomingMessage{
				Message: base.Message{
					Text:    "whats the bots prefix",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "??",
				PermissionLevel: permission.Normal,
			},
			enableNonPrefixCommands: false,
			want:                    nil,
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("[%s]: %s", tc.input.PermissionLevel.Name(), tc.input.Message.Text), func(t *testing.T) {
			db := databasetest.NewFakeDB()
			handler := Handler{db: db, nonPrefixCommandsEnabled: tc.enableNonPrefixCommands}
			got, err := handler.Handle(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Handle() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func flatten[T any](itemGroups ...[]T) []T {
	var items []T
	for _, itemGroup := range itemGroups {
		items = append(items, itemGroup...)
	}
	return items
}

func singleTestCase(tc testCase) []testCase { return []testCase{tc} }

// testCasesWithSameOutput generates test cases that have different message texts
// but are expected to have the same response.
func testCasesWithSameOutput(msgs []string, tc testCase) []testCase {
	var testCases []testCase
	for _, msg := range msgs {
		input := base.IncomingMessage{}
		if err := copier.CopyWithOption(&input, &tc.input, copier.Option{DeepCopy: true}); err != nil {
			panic(err)
		}
		input.Message.Text = msg

		msgTestCase := testCase{}
		if err := copier.CopyWithOption(&msgTestCase, &tc, copier.Option{DeepCopy: true}); err != nil {
			panic(err)
		}
		msgTestCase.input = &input

		testCases = append(testCases, msgTestCase)
	}
	return testCases
}

var (
	savedIVRURL = ivr.BaseURL
)

type fakeExpRandSource struct {
	Value uint64
}

func (s fakeExpRandSource) Uint64() uint64  { return s.Value }
func (s fakeExpRandSource) Seed(val uint64) {}

func setFakes(url string, db *gorm.DB) {
	base.RandReader = bytes.NewBuffer([]byte{3})
	base.RandSource = fakeExpRandSource{Value: uint64(150)}
	ivr.BaseURL = url
	pastebin.FetchPasteURLOverride = url
	twitch.Instance = twitch.NewForTesting(url, db)
}

func resetFakes() {
	base.RandReader = rand.Reader
	base.RandSource = nil
	ivr.BaseURL = savedIVRURL
	pastebin.FetchPasteURLOverride = ""
	twitch.Instance = twitch.NewForTesting(helix.DefaultAPIBaseURL, nil)
}

func joinOtherUser1() error {
	db := databasetest.NewFakeDB()
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
	return err
}

func setRandValueTo0() error {
	base.RandReader = bytes.NewBuffer([]byte{0})
	return nil
}

func setRandValueTo1() error {
	base.RandReader = bytes.NewBuffer([]byte{1})
	return nil
}

func add50PointsToUser1() error {
	db := databasetest.NewFakeDB()
	var user models.User
	result := db.FirstOrCreate(&user, models.User{
		TwitchID:   "user1",
		TwitchName: "user1",
	})
	if result.Error != nil {
		return fmt.Errorf("failed to find/create user: %v", result.Error)
	}
	txn := models.GambaTransaction{
		Game:  "FAKE - TEST",
		User:  user,
		Delta: 50,
	}
	result = db.Create(&txn)
	if result.Error != nil {
		return fmt.Errorf("failed to insert gamba transaction: %v", result.Error)
	}
	return nil
}

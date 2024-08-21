package fun_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/airforce270/airbot/apiclients/bible/bibletest"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/commandtest"
	"github.com/airforce270/airbot/database/databasetest"
	"github.com/airforce270/airbot/permission"
	"github.com/airforce270/airbot/platforms/twitch"
)

func TestFunCommands(t *testing.T) {
	t.Parallel()
	tests := []commandtest.Case{
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$bibleverse Philippians 4:8",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Resources: base.Resources{
					Platform: twitch.NewForTesting(t, "forsen", databasetest.New(t)),
				},
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$bv Philippians 4:8"},
			APIResp:    bibletest.LookupVerseSingleVerse1Resp,
			Want: []*base.Message{
				{
					Text:    "[Philippians 4:8]: Finally, brothers, whatever things are true, whatever things are honorable, whatever things are just, whatever things are pure, whatever things are lovely, whatever things are of good report; if there is any virtue, and if there is any praise, think about these things.\n",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$bibleverse John 3:16",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Resources: base.Resources{
					Platform: twitch.NewForTesting(t, "forsen", databasetest.New(t)),
				},
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$bv John 3:16"},
			APIResp:    bibletest.LookupVerseSingleVerse2Resp,
			Want: []*base.Message{
				{
					Text:    "[John 3:16]: \nFor God so loved the world, that he gave his one and only Son, that whoever believes in him should not perish, but have eternal life.\n\n",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$bibleverse",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
				Resources: base.Resources{
					Platform: twitch.NewForTesting(t, "forsen", databasetest.New(t)),
				},
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$bv"},
			Want: []*base.Message{
				{
					Text:    "Usage: $bibleverse <book> <chapter:verse>",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$cock",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$kok"},
			Want: []*base.Message{
				{
					Text:    "user1's cock is 3 inches long",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$cock someone",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$kok someone"},
			Want: []*base.Message{
				{
					Text:    "someone's cock is 3 inches long",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$fortune",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "Finagle's Creed: Science is true. Don't be misled by facts.",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$iq",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "user1's IQ is 100",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$iq someone",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "someone's IQ is 100",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$ship person1 person2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				func(t testing.TB, r *base.Resources) {
					r.Rand.Reader = bytes.NewBuffer([]byte{95})
				},
			},
			Want: []*base.Message{
				{
					Text:    "person1 and person2 have a 95% compatibility, invite me to the wedding please 😍",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$ship person1 person2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				func(t testing.TB, r *base.Resources) {
					r.Rand.Reader = bytes.NewBuffer([]byte{85})
				},
			},
			Want: []*base.Message{
				{
					Text:    "person1 and person2 have a 85% compatibility, oh 😳",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$ship person1 person2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				func(t testing.TB, r *base.Resources) {
					r.Rand.Reader = bytes.NewBuffer([]byte{70})
				},
			},
			Want: []*base.Message{
				{
					Text:    "person1 and person2 have a 70% compatibility, worth a shot ;)",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$ship person1 person2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				func(t testing.TB, r *base.Resources) {
					r.Rand.Reader = bytes.NewBuffer([]byte{50})
				},
			},
			Want: []*base.Message{
				{
					Text:    "person1 and person2 have a 50% compatibility, it's a toss-up :/",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$ship person1 person2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				func(t testing.TB, r *base.Resources) {
					r.Rand.Reader = bytes.NewBuffer([]byte{30})
				},
			},
			Want: []*base.Message{
				{
					Text:    "person1 and person2 have a 30% compatibility, not sure about this one... :(",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$ship person1 person2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				func(t testing.TB, r *base.Resources) {
					r.Rand.Reader = bytes.NewBuffer([]byte{5})
				},
			},
			Want: []*base.Message{
				{
					Text:    "person1 and person2 have a 5% compatibility, don't even think about it DansGame",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$ship person1",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			Want: []*base.Message{
				{
					Text:    "Usage: $ship <first-person> <second-person>",
					Channel: "user2",
				},
			},
		},
	}

	commandtest.Run(t, tests)
}

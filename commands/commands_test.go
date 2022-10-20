package commands

import (
	"fmt"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/apiclients/ivrtest"
	"github.com/airforce270/airbot/apiclients/twitchtest"
	"github.com/airforce270/airbot/config"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/model"
	"github.com/airforce270/airbot/message"
	"github.com/airforce270/airbot/permission"
	"github.com/airforce270/airbot/platforms/twitch"
	"github.com/airforce270/airbot/testing/fakeserver"

	"github.com/google/go-cmp/cmp"
	"github.com/jinzhu/copier"
	"github.com/nicklaw5/helix/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeFileInfo struct{}

func (f fakeFileInfo) Mode() fs.FileMode  { return 777 }
func (f fakeFileInfo) IsDir() bool        { return false }
func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) Name() string       { return "FakeFile" }
func (f fakeFileInfo) Size() int64        { return 123 }
func (f fakeFileInfo) Sys() any           { return nil }

type testCase struct {
	input     *message.IncomingMessage
	apiResp   string
	runBefore []func() error
	want      []*message.Message
}

func TestCommands(t *testing.T) {
	server := fakeserver.New()
	defer server.Close()

	config.OSReadFile = func(name string) ([]byte, error) {
		return []byte("blahblah"), nil
	}
	config.OSStat = func(name string) (os.FileInfo, error) {
		return fakeFileInfo{}, nil
	}
	config.OSWriteFile = func(name string, data []byte, perm os.FileMode) error {
		return nil
	}

	tests := flatten(
		// admin.go commands
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$join",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: twitchtest.GetChannelInformationResp,
			want: []*message.Message{
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
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$join",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp:   twitchtest.GetChannelInformationResp,
			runBefore: []func() error{joinOtherUser1},
			want: []*message.Message{
				{
					Text:    "Channel user1 is already joined",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$joinother user1",
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
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$joinother user1",
					User:    "user3",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			apiResp: twitchtest.GetChannelInformationResp,
			want: []*message.Message{
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
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$joinother user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			apiResp:   twitchtest.GetChannelInformationResp,
			runBefore: []func() error{joinOtherUser1},
			want: []*message.Message{
				{
					Text:    "Channel user1 is already joined",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$leave",
					User:    "user1",
					Channel: "user1",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Admin,
			},
			apiResp:   twitchtest.GetChannelInformationResp,
			runBefore: []func() error{joinOtherUser1},
			want: []*message.Message{
				{
					Text:    "Successfully left channel.",
					Channel: "user1",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$leave",
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
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$leaveother user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			apiResp:   twitchtest.GetChannelInformationResp,
			runBefore: []func() error{joinOtherUser1},
			want: []*message.Message{
				{
					Text:    "Successfully left channel user1",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$leaveother user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Owner,
			},
			want: []*message.Message{
				{
					Text:    "Bot is not in channel user1",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$leaveother user1",
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
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$help",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: []*message.Message{
				{
					Text:    "For help with a command, use $help <command>. To see available commands, use $commands",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$help join",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: []*message.Message{
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
			input: &message.IncomingMessage{
				Message: message.Message{
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "??",
				PermissionLevel: permission.Normal,
			},
			want: []*message.Message{
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
			input: &message.IncomingMessage{
				Message: message.Message{
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: nil,
		}),

		// echo.go commands
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$commands",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: []*message.Message{
				{
					Text:    "Commands available here: https://github.com/airforce270/airbot#commands",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$TriHard",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: []*message.Message{
				{
					Text:    "TriHard 7",
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
			input: &message.IncomingMessage{
				Message: message.Message{
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.TwitchUsersBannedResp,
			want: []*message.Message{
				{
					Text:    "SeaGrade's ban reason: TOS_INDEFINITE",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$banreason nonbanneduser",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.TwitchUsersNotStreamingResp,
			want: []*message.Message{
				{
					Text:    "xQc is not banned.",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$currentgame",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: twitchtest.GetChannelInformationResp,
			want: []*message.Message{
				{
					Text:    "user1 is currenly playing Science&Technology",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$founders",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.FoundersNormalResp,
			want: []*message.Message{
				{
					Text:    "user1's founders are: FishyyKingyy, eljulidi1337, SamMist, Leochansz, lexieuzumaki7, ContraVz, rott______, DankJuicer, kronikZ____, blemplob",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$founders hasfounders",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.FoundersNormalResp,
			want: []*message.Message{
				{
					Text:    "hasfounders's founders are: FishyyKingyy, eljulidi1337, SamMist, Leochansz, lexieuzumaki7, ContraVz, rott______, DankJuicer, kronikZ____, blemplob",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$founders nofounders",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.FoundersNoneResp,
			want: []*message.Message{
				{
					Text:    "nofounders has no founders",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$founders nofounders404",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.FoundersNone404Resp,
			want: []*message.Message{
				{
					Text:    "nofounders404 has no founders",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$logs xqc forsen",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: []*message.Message{
				{
					Text:    "forsen's logs in xqc's chat: https://logs.ivr.fi/?channel=xqc&username=forsen",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$logs",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			want: []*message.Message{
				{
					Text:    "Usage: $logs <channel> <user>",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$mods",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*message.Message{
				{
					Text:    "user1's mods are: StreamElements, Fossabot, spintto, HNoAce",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$mods otherchannel",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*message.Message{
				{
					Text:    "otherchannel's mods are: StreamElements, Fossabot, spintto, HNoAce",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$mods nomods",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.ModsAndVIPsNoneResp,
			want: []*message.Message{
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
			input: &message.IncomingMessage{
				Message: message.Message{
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: twitchtest.GetChannelInformationResp,
			want: []*message.Message{
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
			input: &message.IncomingMessage{
				Message: message.Message{
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.TwitchUsersVerifiedBotResp,
			want: []*message.Message{
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
			input: &message.IncomingMessage{
				Message: message.Message{
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.TwitchUsersNotVerifiedBotResp,
			want: []*message.Message{
				{
					Text:    "xQc is not a verified bot. ❌",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$vips",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*message.Message{
				{
					Text:    "user1's VIPs are: bakonsword, alyjiahT_T, AVBest, Zaintew_, captkayy, seagrad, Dafkeee",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$vips otherchannel",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*message.Message{
				{
					Text:    "otherchannel's VIPs are: bakonsword, alyjiahT_T, AVBest, Zaintew_, captkayy, seagrad, Dafkeee",
					Channel: "user2",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$vips novips",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			apiResp: ivrtest.ModsAndVIPsNoneResp,
			want: []*message.Message{
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
			setFakes(server.URL())
			database.Instance = newFakeDB()
			for i, f := range tc.runBefore {
				if err := f(); err != nil {
					t.Fatalf("runBefore[%d] func failed: %v", i, err)
				}
			}

			handler := Handler{nonPrefixCommandsEnabled: true}
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
	config.OSStat = os.Stat
	config.OSWriteFile = os.WriteFile
}

func TestCommands_EnableNonPrefixCommands(t *testing.T) {
	tests := []struct {
		input                   *message.IncomingMessage
		enableNonPrefixCommands bool
		want                    []*message.Message
	}{
		{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "whats the bots prefix",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "??",
				PermissionLevel: permission.Normal,
			},
			enableNonPrefixCommands: true,
			want: []*message.Message{
				{
					Text:    "This channel's prefix is ??",
					Channel: "user2",
				},
			},
		},
		{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "whats the bots prefix",
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
			handler := Handler{nonPrefixCommandsEnabled: tc.enableNonPrefixCommands}
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
		input := message.IncomingMessage{}
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

func setFakes(url string) {
	ivr.BaseURL = url
	twitch.Instance = twitch.NewForTesting(url)
}

func resetFakes() {
	ivr.BaseURL = savedIVRURL
	twitch.Instance = twitch.NewForTesting(helix.DefaultAPIBaseURL)
}

func newFakeDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"))
	for _, m := range model.AllModels {
		db.Migrator().DropTable(&m)
	}
	if err != nil {
		panic(err)
	}
	database.Migrate(db)
	return db
}

func joinOtherUser1() error {
	handler := Handler{}
	_, err := handler.Handle(&message.IncomingMessage{
		Message: message.Message{
			Text:    "$joinother user1",
			User:    "user1",
			Channel: "user2",
			Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
		},
		Prefix:          "$",
		PermissionLevel: permission.Owner,
	})
	return err
}

package commands

import (
	"testing"
	"time"

	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/apiclients/ivrtest"
	"github.com/airforce270/airbot/apiclients/twitchtest"
	"github.com/airforce270/airbot/message"
	"github.com/airforce270/airbot/platforms/twitch"
	"github.com/airforce270/airbot/testing/fakeserver"

	"github.com/google/go-cmp/cmp"
	"github.com/jinzhu/copier"
	"github.com/nicklaw5/helix/v2"
)

type testCase struct {
	input   *message.IncomingMessage
	apiResp string
	want    []*message.Message
}

func TestCommands(t *testing.T) {
	server := fakeserver.New()
	server.AddOnClose(resetAPIURLs)
	defer server.Close()
	setAPIURLs(server.URL())

	tests := flatten(
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
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "??",
			},
			want: []*message.Message{
				{
					Text:    "This channel's prefix is ??",
					Channel: "somechannel",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			";prefix",
			"does this bot thingy have one of them prefixes",
			"what is a prefix",
			"forsen prefix",
		}, testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "??",
			},
			want: nil,
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$commands",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			want: []*message.Message{
				{
					Text:    "Commands available here: https://github.com/airforce270/airbot#commands",
					Channel: "somechannel",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$TriHard",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			want: []*message.Message{
				{
					Text:    "TriHard 7",
					Channel: "somechannel",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			"$br",
			"$banreason",
			"$banreason banneduser",
		}, testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: ivrtest.TwitchUsersBannedResp,
			want: []*message.Message{
				{
					Text:    "SeaGrade's ban reason: TOS_INDEFINITE",
					Channel: "somechannel",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$banreason nonbanneduser",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: ivrtest.TwitchUsersNotStreamingResp,
			want: []*message.Message{
				{
					Text:    "xQc is not banned.",
					Channel: "somechannel",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$currentgame",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: twitchtest.GetChannelInformationResp,
			want: []*message.Message{
				{
					Text:    "TwitchDev is currenly playing Science&Technology",
					Channel: "somechannel",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$founders",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: ivrtest.FoundersNormalResp,
			want: []*message.Message{
				{
					Text:    "someone's founders are: FishyyKingyy, eljulidi1337, SamMist, Leochansz, lexieuzumaki7, ContraVz, rott______, DankJuicer, kronikZ____, blemplob",
					Channel: "somechannel",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$founders hasfounders",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: ivrtest.FoundersNormalResp,
			want: []*message.Message{
				{
					Text:    "hasfounders's founders are: FishyyKingyy, eljulidi1337, SamMist, Leochansz, lexieuzumaki7, ContraVz, rott______, DankJuicer, kronikZ____, blemplob",
					Channel: "somechannel",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$founders nofounders",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: ivrtest.FoundersNoneResp,
			want: []*message.Message{
				{
					Text:    "nofounders has no founders",
					Channel: "somechannel",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$founders nofounders404",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: ivrtest.FoundersNone404Resp,
			want: []*message.Message{
				{
					Text:    "nofounders404 has no founders",
					Channel: "somechannel",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$mods",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*message.Message{
				{
					Text:    "someone's mods are: StreamElements, Fossabot, spintto, HNoAce",
					Channel: "somechannel",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$mods otherchannel",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*message.Message{
				{
					Text:    "otherchannel's mods are: StreamElements, Fossabot, spintto, HNoAce",
					Channel: "somechannel",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$mods nomods",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: ivrtest.ModsAndVIPsNoneResp,
			want: []*message.Message{
				{
					Text:    "nomods has no mods",
					Channel: "somechannel",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			"$title",
			"$title otherchannel",
		}, testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: twitchtest.GetChannelInformationResp,
			want: []*message.Message{
				{
					Text:    "TwitchDev's title: TwitchDevMonthlyUpdate//May6,2021",
					Channel: "somechannel",
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
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: ivrtest.TwitchUsersVerifiedBotResp,
			want: []*message.Message{
				{
					Text:    "iP0G is a verified bot. ✅",
					Channel: "somechannel",
				},
			},
		}),
		testCasesWithSameOutput([]string{
			"$verifiedbot notverified",
			"$vb notverified",
		}, testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: ivrtest.TwitchUsersNotVerifiedBotResp,
			want: []*message.Message{
				{
					Text:    "xQc is not a verified bot. ❌",
					Channel: "somechannel",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$vips",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*message.Message{
				{
					Text:    "someone's VIPs are: bakonsword, alyjiahT_T, AVBest, Zaintew_, captkayy, seagrad, Dafkeee",
					Channel: "somechannel",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$vips otherchannel",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: ivrtest.ModsAndVIPsModsAndVIPsResp,
			want: []*message.Message{
				{
					Text:    "otherchannel's VIPs are: bakonsword, alyjiahT_T, AVBest, Zaintew_, captkayy, seagrad, Dafkeee",
					Channel: "somechannel",
				},
			},
		}),
		singleTestCase(testCase{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$vips novips",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			apiResp: ivrtest.ModsAndVIPsNoneResp,
			want: []*message.Message{
				{
					Text:    "novips has no VIPs",
					Channel: "somechannel",
				},
			},
		}),
	)

	for _, tc := range tests {
		if tc.apiResp != "" {
			server.Resp = tc.apiResp
		}
		t.Run(tc.input.Message.Text, func(t *testing.T) {
			handler := Handler{nonPrefixCommandsEnabled: true}
			got, err := handler.Handle(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Handle() diff (-want +got):\n%s", diff)
			}
		})
		server.Reset()
	}
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
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "??",
			},
			enableNonPrefixCommands: true,
			want: []*message.Message{
				{
					Text:    "This channel's prefix is ??",
					Channel: "somechannel",
				},
			},
		},
		{
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "whats the bots prefix",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "??",
			},
			enableNonPrefixCommands: false,
			want:                    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.input.Message.Text, func(t *testing.T) {
			handler := NewHandler(tc.enableNonPrefixCommands)
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

func setAPIURLs(url string) {
	ivr.BaseURL = url
	twitch.Instance = twitch.NewForTesting(url)
}

func resetAPIURLs() {
	ivr.BaseURL = savedIVRURL
	twitch.Instance = twitch.NewForTesting(helix.DefaultAPIBaseURL)
}

// Package commandtest provides helpers for testing commands.
package commandtest

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/airforce270/airbot/apiclients/bible"
	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/apiclients/kick"
	"github.com/airforce270/airbot/apiclients/seventv"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/cache/cachetest"
	"github.com/airforce270/airbot/commands"
	"github.com/airforce270/airbot/database/databasetest"
	"github.com/airforce270/airbot/platforms/twitch"
	"github.com/airforce270/airbot/testing/fakeserver"
	"github.com/google/go-cmp/cmp"
)

// Case is a test case for running command tests.
type Case struct {
	Input      base.IncomingMessage
	Platform   Platform
	OtherTexts []string
	ApiResp    string
	ApiResps   []string
	ConfigData string
	RunBefore  []SetupFunc
	RunAfter   []TeardownFunc
	Want       []*base.Message
}

func Run(t *testing.T, tests []Case) {
	t.Helper()
	for _, tc := range buildTestCases(tests) {
		tc := tc
		t.Run(fmt.Sprintf("[%s] %s", tc.input.PermissionLevel.Name(), tc.input.Message.Text), func(t *testing.T) {
			t.Helper()
			t.Parallel()
			server := fakeserver.New()
			defer server.Close()

			server.Resps = tc.apiResps

			db := databasetest.New(t)
			cdb := cachetest.NewInMemory()

			var platform base.Platform
			switch tc.platform {
			case TwitchPlatform:
				platform = twitch.NewForTesting(server.URL(), db)
			default:
				t.Fatal("Platform must be set.")
			}

			resources := base.Resources{
				Platform: platform,
				DB:       db,
				Cache:    cdb,
				AllPlatforms: map[string]base.Platform{
					platform.Name(): platform,
				},
				NewConfigSource: func() (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(tc.configData)), nil
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

			for _, f := range tc.runBefore {
				f(t, &resources)
			}

			tc.input.Resources = resources

			handler := commands.NewHandlerForTest(db, cdb, resources.AllPlatforms, resources.NewConfigSource, resources.Rand, resources.Clients)
			got, err := handler.Handle(&tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, f := range tc.runAfter {
				f(t)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Handle() diff (-want +got):\n%s", diff)
			}
		})
	}
}

// SetupFunc is a function to be run before a test case runs.
type SetupFunc func(testing.TB, *base.Resources)

// TeardownFunc is a function to be run after a test case runs.
type TeardownFunc func(testing.TB)

// Platform indicates which platform to consider the message to be sent from.
type Platform uint8

const (
	TwitchPlatform Platform = iota
)

// builtCase is a built test case for running command tests.
type builtCase struct {
	input      base.IncomingMessage
	platform   Platform
	apiResps   []string
	configData string
	runBefore  []SetupFunc
	runAfter   []TeardownFunc
	want       []*base.OutgoingMessage
}

func buildTestCases(tcs []Case) []builtCase {
	var builtCases []builtCase

	for _, tc := range tcs {
		texts := append([]string{tc.Input.Message.Text}, tc.OtherTexts...)
		var apiResps []string
		if tc.ApiResp != "" {
			apiResps = append(apiResps, tc.ApiResp)
		}
		apiResps = append(apiResps, tc.ApiResps...)
		for _, text := range texts {
			builtCase := builtCase{
				input:      tc.Input,
				platform:   tc.Platform,
				apiResps:   apiResps,
				configData: tc.ConfigData,
				runBefore:  tc.RunBefore,
				runAfter:   tc.RunAfter,
			}
			builtCase.input.Message.Text = text
			for _, want := range tc.Want {
				builtCase.want = append(builtCase.want, &base.OutgoingMessage{Message: *want})
			}
			builtCases = append(builtCases, builtCase)
		}
	}

	return builtCases
}

type fakeExpRandSource struct {
	Value uint64
}

func (s fakeExpRandSource) Uint64() uint64  { return s.Value }
func (s fakeExpRandSource) Seed(val uint64) {}

// Package commandtest provides helpers for testing commands.
package commandtest

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/airforce270/airbot/apiclients/bible"
	"github.com/airforce270/airbot/apiclients/ivr"
	"github.com/airforce270/airbot/apiclients/kick"
	"github.com/airforce270/airbot/apiclients/pastebin"
	"github.com/airforce270/airbot/apiclients/seventv"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/cache/cachetest"
	"github.com/airforce270/airbot/commands"
	"github.com/airforce270/airbot/database/databasetest"
	"github.com/airforce270/airbot/platforms/twitch"
	"github.com/airforce270/airbot/testing/fakeserver"
	"github.com/google/go-cmp/cmp"
	"github.com/nicklaw5/helix/v2"
	"gorm.io/gorm"
)

// Case is a test case for running command tests.
type Case struct {
	Input      base.IncomingMessage
	Platform   Platform
	OtherTexts []string
	ApiResp    string
	ApiResps   []string
	RunBefore  []SetupFunc
	RunAfter   []TeardownFunc
	Want       []*base.Message
}

func Run(t *testing.T, tests []Case) {
	t.Helper()
	for _, tc := range buildTestCases(tests) {
		t.Run(fmt.Sprintf("[%s] %s", tc.input.PermissionLevel.Name(), tc.input.Message.Text), func(t *testing.T) {
			t.Helper()
			server := fakeserver.New()
			defer server.Close()

			server.Resps = tc.apiResps

			db := databasetest.NewFakeDB(t)
			cdb := cachetest.NewInMemory()

			setFakes(server.URL(), db)

			var platform base.Platform
			switch tc.platform {
			case TwitchPlatform:
				platform = twitch.NewForTesting(server.URL(), databasetest.NewFakeDB(t))
			default:
				t.Fatal("Platform must be set.")
			}

			resources := base.Resources{
				Platform: platform,
				DB:       db,
				Cache:    cdb,
				Rand: base.RandResources{
					Reader: bytes.NewBuffer([]byte{3}),
					Source: fakeExpRandSource{Value: uint64(150)},
				},
			}

			for _, f := range tc.runBefore {
				f(t, &resources)
			}

			tc.input.Resources = resources

			handler := commands.NewHandler(db, cdb, resources.Rand)
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
			resetFakes()
			server.Reset()
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
	input     base.IncomingMessage
	platform  Platform
	apiResps  []string
	runBefore []SetupFunc
	runAfter  []TeardownFunc
	want      []*base.OutgoingMessage
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
				input:     tc.Input,
				platform:  tc.Platform,
				apiResps:  apiResps,
				runBefore: tc.RunBefore,
				runAfter:  tc.RunAfter,
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
	bible.BaseURL = url
	ivr.BaseURL = url
	kick.BaseURL = url
	pastebin.FetchPasteURLOverride = url
	seventv.BaseURL = url
	twitch.SetInstance(twitch.NewForTesting(url, db))
}

func resetFakes() {
	bible.BaseURL = savedBibleURL
	ivr.BaseURL = savedIVRURL
	kick.BaseURL = savedKickURL
	pastebin.FetchPasteURLOverride = ""
	seventv.BaseURL = saved7TVURL
	twitch.SetInstance(twitch.NewForTesting(helix.DefaultAPIBaseURL, nil))
}

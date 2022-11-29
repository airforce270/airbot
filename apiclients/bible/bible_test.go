package bible

import (
	"testing"

	"github.com/airforce270/airbot/apiclients/bible/bibletest"
	"github.com/airforce270/airbot/testing/fakeserver"

	"github.com/google/go-cmp/cmp"
)

var (
	originalBaseURL = BaseURL
)

func TestFetchUser(t *testing.T) {
	server := fakeserver.New()
	server.AddOnClose(func() { originalBaseURL = BaseURL })
	defer server.Close()
	BaseURL = server.URL()

	tests := []struct {
		desc    string
		useResp string
		want    *GetVersesResponse
	}{
		{
			desc:    "single verse",
			useResp: bibletest.LookupVerseSingleVerse1Resp,
			want: &GetVersesResponse{
				Reference: "Philippians 4:8",
				Verses: []Verse{
					{
						BookID:   "PHP",
						BookName: "Philippians",
						Chapter:  4,
						Verse:    8,
						Text:     "Finally, brothers, whatever things are true, whatever things are honorable, whatever things are just, whatever things are pure, whatever things are lovely, whatever things are of good report; if there is any virtue, and if there is any praise, think about these things.\n",
					},
				},
				Text:            "Finally, brothers, whatever things are true, whatever things are honorable, whatever things are just, whatever things are pure, whatever things are lovely, whatever things are of good report; if there is any virtue, and if there is any praise, think about these things.\n",
				TranslationID:   "web",
				TranslationName: "World English Bible",
				TranslationNote: "Public Domain",
			},
		},
	}

	for _, tc := range tests {
		server.Resp = tc.useResp
		t.Run(tc.desc, func(t *testing.T) {
			got, err := FetchVerses("Philippians 4:8")
			if err != nil {
				t.Fatalf("FetchVerse() unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("FetchVerse() diff (-want +got):\n%s", diff)
			}
		})
		server.Reset()
	}
}

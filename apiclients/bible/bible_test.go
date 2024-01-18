package bible_test

import (
	"testing"

	"github.com/airforce270/airbot/apiclients/bible"
	"github.com/airforce270/airbot/apiclients/bible/bibletest"
	"github.com/airforce270/airbot/testing/fakeserver"

	"github.com/google/go-cmp/cmp"
)

func TestFetchUser(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc    string
		useResp string
		want    *bible.GetVersesResponse
	}{
		{
			desc:    "single verse",
			useResp: bibletest.LookupVerseSingleVerse1Resp,
			want: &bible.GetVersesResponse{
				Reference: "Philippians 4:8",
				Verses: []bible.Verse{
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
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			server := fakeserver.New()
			defer server.Close()
			server.Resps = []string{tc.useResp}

			client := bible.NewClient(server.URL(t).String())
			got, err := client.FetchVerses("Philippians 4:8")
			if err != nil {
				t.Fatalf("FetchVerses() unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("FetchVerses() diff (-want +got):\n%s", diff)
			}
		})
	}
}

package bible

import (
	"testing"

	"github.com/airforce270/airbot/apiclients/bibletest"
	"github.com/airforce270/airbot/testing/fakeserver"

	"github.com/google/go-cmp/cmp"
)

var (
	originalBibleBaseURL = BaseURL
)

func TestFetchUser(t *testing.T) {
	server := fakeserver.New()
	server.AddOnClose(func() { originalBibleBaseURL = BaseURL })
	defer server.Close()
	BaseURL = server.URL()

	tests := []struct {
		desc    string
		useResp string
		want    Verse
	}{
		{
			desc:    "single verse",
			useResp: bibletest.LookupVerseSingleVerseResp,
			want: Verse{
				BookID:   "PHP",
				BookName: "Philippians",
				Chapter:  4,
				Verse:    8,
				Text:     "Finally, brothers, whatever things are true, whatever things are honorable, whatever things are just, whatever things are pure, whatever things are lovely, whatever things are of good report; if there is any virtue, and if there is any praise, think about these things.\n",
			},
		},
	}

	for _, tc := range tests {
		server.Resp = tc.useResp
		t.Run(tc.desc, func(t *testing.T) {
			got, err := FetchVerse("Philippians 4:8")
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

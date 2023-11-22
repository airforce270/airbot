package pastebin

import (
	"testing"

	"github.com/airforce270/airbot/apiclients/pastebin/pastebintest"
	"github.com/airforce270/airbot/testing/fakeserver"

	"github.com/google/go-cmp/cmp"
)

func TestFetchPaste(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc    string
		useResp string
		want    Paste
	}{
		{
			desc:    "single-line",
			useResp: pastebintest.SingleLineFetchPasteResp,
			want:    Paste([]string{"line1"}),
		},
		{
			desc:    "multi-line",
			useResp: pastebintest.MultiLineFetchPasteResp,
			want:    Paste([]string{"line1", "line2", "line3"}),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			server := fakeserver.New()
			defer server.Close()
			server.Resps = []string{tc.useResp}
			got, err := FetchPaste(server.URL())
			if err != nil {
				t.Fatalf("FetchPaste() unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("FetchPaste() diff (-want +got):\n%s", diff)
			}
		})
	}
}

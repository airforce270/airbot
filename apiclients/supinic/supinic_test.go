package supinic

import (
	"testing"

	"github.com/airforce270/airbot/apiclients/supinic/supinictest"
	"github.com/airforce270/airbot/testing/fakeserver"
)

func TestUpdateBotActivity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc    string
		useResp string
		wantErr bool
	}{
		{
			desc:    "success",
			useResp: supinictest.UpdateBotActivitySuccessResp,
			wantErr: false,
		},
		{
			desc:    "streaming user",
			useResp: supinictest.UpdateBotActivityFailureResp,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			server := fakeserver.New()
			defer server.Close()
			server.Resps = []string{tc.useResp}
			client := NewClientForTesting(server.URL(t).String())

			err := client.updateBotActivity()
			if !tc.wantErr && err != nil {
				t.Fatalf("updateBotActivity() unexpected error: %v", err)
			}

			if tc.wantErr && err == nil {
				t.Errorf("updateBotActivity() want error, but returned none")
			}
		})
	}
}

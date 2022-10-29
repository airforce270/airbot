package supinic

import (
	"testing"

	"github.com/airforce270/airbot/apiclients/supinictest"
	"github.com/airforce270/airbot/testing/fakeserver"
)

func TestUpdateBotActivity(t *testing.T) {
	server := fakeserver.New()
	defer server.Close()

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
		server.Resp = tc.useResp
		t.Run(tc.desc, func(t *testing.T) {
			client := NewClientForTesting(server.URL())

			err := client.updateBotActivity()
			if !tc.wantErr && err != nil {
				t.Fatalf("updateBotActivity() unexpected error: %v", err)
			}

			if tc.wantErr && err == nil {
				t.Errorf("updateBotActivity() want error, but returned none")
			}
		})
		server.Reset()
	}
}

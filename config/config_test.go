package config

import (
	_ "embed"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

//go:embed config_example.toml
var configExample []byte

func TestParse(t *testing.T) {
	want := &Config{
		LogIncoming: true,
		LogOutgoing: true,
		Platforms: PlatformConfig{
			Kick: KickConfig{
				JA3:       "",
				UserAgent: "",
			},
			Twitch: TwitchConfig{
				Enabled:     true,
				Username:    "",
				ClientID:    "",
				AccessToken: "",
				Owners:      []string{""},
			},
		},
		Supinic: SupinicConfig{
			UserID:        "not-required-to-run-bot",
			APIKey:        "you-can-safely-leave-this-as-is",
			ShouldPingAPI: false,
		},
	}

	got, err := parse(configExample)
	if err != nil {
		t.Fatalf("parse() unexpected error: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parse() diff (-want +got):\n%s", diff)
	}
}

func TestSupinicConfigIsConfigured(t *testing.T) {
	tests := []struct {
		input SupinicConfig
		want  bool
	}{
		{
			input: SupinicConfig{
				UserID:        "surely-a-real-user-id",
				APIKey:        "surely-a-real-api-key",
				ShouldPingAPI: true,
			},
			want: true,
		},
		{
			input: SupinicConfig{
				UserID:        "",
				APIKey:        "",
				ShouldPingAPI: true,
			},
			want: false,
		},
		{
			input: SupinicConfig{
				UserID:        placeholderSupinicUserID,
				APIKey:        placeholderSupinicAPIKey,
				ShouldPingAPI: true,
			},
			want: false,
		},
	}

	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := tc.input.IsConfigured(); got != tc.want {
				t.Errorf("SupinicConfig.IsConfigured() = %t, want %t", got, tc.want)
			}
		})
	}
}

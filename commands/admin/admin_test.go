package admin

import (
	"bytes"
	"os"
	"testing"

	"github.com/airforce270/airbot/apiclients/kick"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/config"
	"github.com/pelletier/go-toml/v2"
)

func TestReloadConfig(t *testing.T) {
	const want = "something-specific"
	config.OSReadFile = func(name string) ([]byte, error) {
		var buf bytes.Buffer
		enc := toml.NewEncoder(&buf)
		if err := enc.Encode(&config.Config{
			Platforms: config.PlatformConfig{
				Kick: config.KickConfig{
					UserAgent: want,
				},
			},
		}); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	defer func() { config.OSReadFile = os.ReadFile }()
	msg := base.IncomingMessage{
		Message: base.Message{
			Channel: "someone",
		},
	}

	_, err := reloadConfig(&msg, nil)
	if err != nil {
		t.Fatal(err)
	}

	if got := *kick.UserAgent.Load(); got != want {
		t.Errorf("reloadConfig() value = %q, want %q", got, want)
	}
}

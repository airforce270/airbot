// Package ivrtest provides helpers for testing connections to the Supinic API.
package supinictest

import _ "embed"

var (
	//go:embed update_bot_activity/success.json
	UpdateBotActivitySuccessResp string
	//go:embed update_bot_activity/failure.json
	UpdateBotActivityFailureResp string
)

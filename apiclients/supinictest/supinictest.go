// Package ivrtest provides helpers for testing connections to the Supinic API.
package supinictest

const (
	UpdateBotActivitySuccessResp = `{"statusCode":200,"timestamp":1667031962798,"data":{"success":true},"error":null}`
	UpdateBotActivityFailureResp = `{"statusCode":401,"timestamp":1667031962798,"data":{"success":false},"error":"Authorization failed"}`
)

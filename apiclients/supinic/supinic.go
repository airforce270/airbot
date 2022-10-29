// Package supinic contains a client for the Supinic API.
// https://supinic.com/api/
package supinic

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Client is a Supinic API client.
type Client struct {
	// userID is the Supinic User ID of the bot.
	// https://supinic.com/user/auth-key
	userID string
	// apiKey is an authentication key for the bot.
	// https://supinic.com/user/auth-key
	apiKey string
	// baseURL is the base URL of the API to use.
	baseURL string
	// h is an HTTP client to make requests with.
	h http.Client
}

const pingInterval = time.Duration(15) * time.Minute

// StartPinging starts a background task to ping the Supinic API regularly
// to make sure the API knows the bot is still online.
// This function blocks and should be run within a goroutine.
func (c *Client) StartPinging() {
	for {
		go c.pingAPI()
		time.Sleep(pingInterval)
	}
}

func (c *Client) pingAPI() {
	if err := c.updateBotActivity(); err != nil {
		log.Printf("Failed to ping Supinic API: %v", err)
	}
}

// baseAPIResponse contains the common fields returned by the IVR API.
// This struct should be embedded in a API-call specific struct.
// The actual response data will be contained in the Data field
// and can (should) overridden with a more specific type.
type baseAPIResponse struct {
	// StatusCode is the status code of the response.
	StatusCode int `json:"statusCode"`
	// TimestampMs is the timestamp in milliseconds when the response was sent.
	TimestampMs int64 `json:"timestamp"`
	// Data is the response data itself.
	Data any `json:"data"`
	// Error is the error, if any.
	Error string `json:"error"`
}

// updateBotActivity updates the Supinic API to indicate this bot is online.
// https://supinic.com/api/#api-Bot_Program-updateBotActivity
func (c *Client) updateBotActivity() error {
	apiResp, err := c.put("bot-program/bot/active")
	if err != nil {
		return err
	}

	resp := struct {
		baseAPIResponse
		Data struct {
			Success bool `json:"success"`
		} `json:"data"`
	}{}
	if err := json.Unmarshal(apiResp, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal UpdateBotActivity response (%q): %w", apiResp, err)
	}

	if !resp.Data.Success {
		return fmt.Errorf("supinic UpdateBotActivity returned success=false, resp: %q", apiResp)
	}

	return nil
}

func (c *Client) call(method, path string, body io.Reader) ([]byte, error) {
	reqURL, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL (%q, %q): %w", c.baseURL, path, err)
	}
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s:%s", c.userID, c.apiKey))

	resp, err := c.h.Do(req)
	if err != nil {
		return nil, fmt.Errorf("supinic API call failed (req: %v): %w", req, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from supinic API (req: %v): %w", resp, err)
	}
	return respBody, err
}

func (c *Client) put(path string) ([]byte, error) {
	return c.call(http.MethodPut, path, nil)
}

// NewClient creates a new Client.
func NewClient(userID, apiKey string) *Client {
	return &Client{
		userID:  userID,
		apiKey:  apiKey,
		baseURL: "https://supinic.com/api/",
		h:       http.Client{},
	}
}

// NewClientForTesting creates a new Client for testing.
func NewClientForTesting(baseURL string) *Client {
	return &Client{
		userID:  "fake-user-id",
		apiKey:  "fake-api-key",
		baseURL: baseURL,
		h:       http.Client{},
	}
}

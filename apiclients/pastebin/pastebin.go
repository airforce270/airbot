// Package pastebin handles interactions with the Pastebin API.
package pastebin

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// NewClient creates a new Pastebin API client.
// fetchPasteURLOverride is optional and should only be set in test.
func NewClient(fetchPasteURLOverride string) *Client {
	return &Client{fetchPasteURLOverride: fetchPasteURLOverride}
}

// Client is a client for the Pastebin API.
type Client struct {
	fetchPasteURLOverride string
}

// FetchPaste fetches a paste, given a pastebin URL.
// Example: https://pastebin.com/raw/B7TBjQEy
func (c *Client) FetchPaste(pasteURL string) (Paste, error) {
	reqURL := pasteURL
	if c.fetchPasteURLOverride != "" {
		reqURL = c.fetchPasteURLOverride
	}

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch paste from pastebin (URL:%s): %w", reqURL, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from Pastebin API (URL:%s): %v", reqURL, resp)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from Pastebin API: %w", err)
	}
	lines := strings.Split(string(body), "\n")

	return Paste(lines), nil
}

// Paste represents a pastebin paste.
type Paste []string

// Values returns the paste's values.
func (p Paste) Values() []string { return []string(p) }

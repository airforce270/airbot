// Package pastebin handles interactions with the Pastebin API.
package pastebin

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Paste represents a pastebin paste.
type Paste []string

func (p Paste) Values() []string { return []string(p) }

// FetchPasteURLOverride, if set, overrides the URL called, for testing.
var FetchPasteURLOverride = ""

// FetchPaste fetches a paste, given a pastebin URL.
// Example: https://pastebin.com/raw/B7TBjQEy
func FetchPaste(pasteURL string) (Paste, error) {
	reqURL := pasteURL
	if FetchPasteURLOverride != "" {
		reqURL = FetchPasteURLOverride
	}

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, err
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

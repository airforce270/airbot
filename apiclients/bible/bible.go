// Package bible provides an API client to the bible-api.com API.
package bible

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Base URL for API requests. Should only be changed for testing.
var BaseURL = "https://bible-api.com"

// GetVersesResponse represents the response from the Bible API for a GetVerses request.
// https://bible-api.com/
type GetVersesResponse struct {
	// Reference is the specific reference to the verse.
	// ex: "John 3:16"
	Reference string `json:"reference"`
	// Verses contains the specific verses returned.
	Verses []Verse `json:"verses"`
	// Text is the combined text of the returned verses.
	// ex: "For God so loved the world, that he gave his one and only Son, that whoever believes in him should not perish, but have eternal life."
	Text string `json:"text"`
	// TranslationID is the short identifier of the translation used.
	// ex: "web"
	TranslationID string `json:"translation_id"`
	// TranslationName is the human-readable name of the translation used.
	// ex: "World English Bible"
	TranslationName string `json:"translation_name"`
	// TranslationNote contains notes about the translation used.
	// ex: "Public Domain"
	TranslationNote string `json:"translation_note"`
}

// Verse represents a specific verse from the Bible.
type Verse struct {
	// BookID is the short ID of the book.
	// ex: "JHN"
	BookID string `json:"book_id"`
	// BookName is the human-readable name of the book.
	// ex: "John"
	BookName string `json:"book_name"`
	// Chapter is the chapter of the verse.
	// ex: 3
	Chapter int `json:"chapter"`
	// Verse is the verse number.
	// ex: 16
	Verse int `json:"verse"`
	// Text is the text of the verse.
	// ex: "For God so loved the world, that he gave his one and only Son, that whoever believes in him should not perish, but have eternal life."
	Text string `json:"text"`
}

func get(reqURL string) (respBody []byte, err error) {
	httpResp, err := http.Get(reqURL)
	if err != nil {
		return nil, err
	}
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from Bible API (URL:%s): %v", reqURL, httpResp)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from Bible API: %w", err)
	}

	return body, nil
}

func FetchVerses(verse string) (*GetVersesResponse, error) {
	body, err := get(fmt.Sprintf("%s/%s", BaseURL, url.QueryEscape(verse)))
	if err != nil {
		return nil, err
	}

	resp := GetVersesResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from Bible API: %w", err)
	}

	return &resp, nil
}

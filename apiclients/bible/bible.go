// Package bible provides an API client to the bible-api.com API.
package bible

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// Base URL for API requests. Should only be changed for testing.
var BaseURL = "https://bible-api.com"

type GetVerseResponse struct {
	Reference       string  `json:"reference"`
	Verses          []Verse `json:"verses"`
	Text            string  `json:"text"`
	TranslationID   string  `json:"translation_id"`
	TranslationName string  `json:"translation_name"`
	TranslationNote string  `json:"translation_note"`
}

type Verse struct {
	BookID   string `json:"book_id"`
	BookName string `json:"book_name"`
	Chapter  int    `json:"chapter"`
	Verse    int    `json:"verse"`
	Text     string `json:"text"`
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

func FetchVerse(verse string) (Verse, error) {
	body, err := get(fmt.Sprintf("%s/%s", BaseURL, url.QueryEscape(verse)))
	if err != nil {
		return Verse{}, err
	}

	resp := GetVerseResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return Verse{}, fmt.Errorf("failed to unmarshal response from Bible API: %w", err)
	}

	if len(resp.Verses) == 0 {
		return Verse{}, fmt.Errorf("no verses returned: %v", resp)
	}
	if len(resp.Verses) > 1 {
		log.Printf("Matched more than 1 verse: %v", resp)
	}

	return resp.Verses[0], nil
}

// Package bibletest provides helpers for testing connections to the Bible API.
package bibletest

import _ "embed"

var (
	//go:embed lookup_verse/single_verse_1.json
	LookupVerseSingleVerse1Resp string
	//go:embed lookup_verse/single_verse_2.json
	LookupVerseSingleVerse2Resp string
)

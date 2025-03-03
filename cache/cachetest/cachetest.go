// Package cachetest provides a fake cache for testing.
package cachetest

import (
	"testing"

	"github.com/airforce270/airbot/cache"
	"github.com/airforce270/airbot/database"
)

// NewSQLite creates a new cache for test.
func NewSQLite(t *testing.T, queries *database.Queries) cache.Cache {
	t.Helper()
	c, err := cache.NewSQLite(queries)
	if err != nil {
		t.Fatalf("Failed to create cache for test: %v", err)
	}
	return &c
}

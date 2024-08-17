// Package cachetest provides a fake cache for testing.
package cachetest

import (
	"testing"

	"github.com/airforce270/airbot/cache"
	"gorm.io/gorm"
)

// NewSQLite creates a new cache for test.
func NewSQLite(t *testing.T, db *gorm.DB) cache.Cache {
	t.Helper()
	c, err := cache.NewSQLite(db)
	if err != nil {
		t.Fatalf("Failed to create cache for test: %v", err)
	}
	return &c
}

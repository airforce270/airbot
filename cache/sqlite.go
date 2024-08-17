package cache

import (
	"fmt"
	"time"

	"github.com/airforce270/airbot/database/models"
	"gorm.io/gorm"
)

// NewSQLite creates a new SQLite-backed Cache.
func NewSQLite(db *gorm.DB) (SQLite, error) {
	return SQLite{db}, nil
}

// SQLite implements Cache for a SQLite database.
type SQLite struct {
	db *gorm.DB
}

func (v *SQLite) StoreBool(key string, value bool) error {
	item := models.CacheBoolItem{
		Key:   key,
		Value: value,
	}

	if err := v.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to store %s=%v: %w", key, value, err)
	}

	return nil
}

func (v *SQLite) StoreExpiringBool(key string, value bool, expiration time.Duration) error {
	item := models.CacheBoolItem{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(expiration),
	}

	if err := v.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to store %q=%t: %w", key, value, err)
	}

	return nil
}

func (v *SQLite) FetchBool(key string) (bool, error) {
	var item models.CacheBoolItem

	if err := v.db.First(&item, "key = ?", key).Error; err != nil {
		return false, fmt.Errorf("failed to fetch %q: %w", key, err)
	}

	if !item.ExpiresAt.IsZero() && item.ExpiresAt.Before(time.Now()) {
		return false, nil
	}

	return item.Value, nil
}

func (v *SQLite) StoreString(key, value string) error {
	item := models.CacheStringItem{
		Key:   key,
		Value: value,
	}

	if err := v.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to store %q=%q: %w", key, value, err)
	}

	return nil
}

func (v *SQLite) StoreExpiringString(key, value string, expiration time.Duration) error {
	item := models.CacheStringItem{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(expiration),
	}

	if err := v.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to store %q=%q: %w", key, value, err)
	}

	return nil
}

func (v *SQLite) FetchString(key string) (string, error) {
	var item models.CacheStringItem

	if err := v.db.First(&item, "key = ?", key).Error; err != nil {
		return "", fmt.Errorf("failed to fetch %q: %w", key, err)
	}

	if !item.ExpiresAt.IsZero() && item.ExpiresAt.Before(time.Now()) {
		return "", nil
	}

	return item.Value, nil
}

package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/utils/ptrs"
)

// NewSQLite creates a new SQLite-backed Cache.
func NewSQLite(queries *database.Queries) (SQLite, error) {
	return SQLite{queries}, nil
}

// SQLite implements Cache for a SQLite database.
type SQLite struct {
	queries *database.Queries
}

func (v *SQLite) StoreBool(ctx context.Context, key string, value bool) error {
	return v.StoreExpiringBool(ctx, key, value, -1 /* expiration */)
}

func (v *SQLite) StoreExpiringBool(ctx context.Context, key string, value bool, expiration time.Duration) error {
	item := database.UpsertCacheBoolItemParams{
		Keyy: ptrs.StringNil(key),
	}
	if value {
		item.Value = ptrs.TrueFloat
	} else {
		item.Value = ptrs.FalseFloat
	}
	if expiration >= 0 {
		item.ExpiresAt = ptrs.Ptr(time.Now().Add(expiration))
	}

	if err := v.queries.UpsertCacheBoolItem(ctx, item); err != nil {
		return fmt.Errorf("failed to store %s=%v: %w", key, value, err)
	}

	return nil
}

func (v *SQLite) FetchBool(ctx context.Context, key string) (bool, error) {
	item, err := v.queries.SelectCacheBoolItem(ctx, ptrs.StringNil(key))
	if err != nil {
		return false, fmt.Errorf("failed to fetch %q: %w", key, err)
	}

	if !item.ExpiresAt.IsZero() && item.ExpiresAt.Before(time.Now()) {
		return false, nil
	}

	return item.Value == ptrs.TrueFloat, nil
}

func (v *SQLite) StoreString(ctx context.Context, key, value string) error {
	return v.StoreExpiringString(ctx, key, value, -1 /* expiration */)
}

func (v *SQLite) StoreExpiringString(ctx context.Context, key, value string, expiration time.Duration) error {
	item := database.UpsertCacheStringItemParams{
		Keyy:  ptrs.StringNil(key),
		Value: ptrs.StringNil(value),
	}
	if expiration >= 0 {
		item.ExpiresAt = ptrs.Ptr(time.Now().Add(expiration))
	}

	if err := v.queries.UpsertCacheStringItem(ctx, item); err != nil {
		return fmt.Errorf("failed to store %s=%v: %w", key, value, err)
	}

	return nil
}

func (v *SQLite) FetchString(ctx context.Context, key string) (string, error) {
	item, err := v.queries.SelectCacheStringItem(ctx, ptrs.StringNil(key))
	if err != nil {
		return "", fmt.Errorf("failed to fetch %q: %w", key, err)
	}

	if !item.ExpiresAt.IsZero() && item.ExpiresAt.Before(time.Now()) {
		return "", nil
	}

	if item.Value == nil {
		return "", nil
	}

	return *item.Value, nil
}

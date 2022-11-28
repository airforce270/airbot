// Package cachetest provides a fake cache for testing.
package cachetest

import (
	"sync"
	"time"
)

// NewInMemoryCache creates a new InMemoryCache for test.
func NewInMemoryCache() *InMemoryCache {
	m := map[string]any{}
	c := InMemoryCache{d: m}
	return &c
}

// InMemoryCache implements a simple map-based cache for testing.
type InMemoryCache struct {
	d   map[string]any
	mtx sync.Mutex
}

func (c *InMemoryCache) StoreBool(key string, value bool) error {
	c.store(key, value)
	return nil
}
func (c *InMemoryCache) StoreExpiringBool(key string, value bool, expiration time.Duration) error {
	c.store(key, value)
	return nil
}
func (c *InMemoryCache) FetchBool(key string) (bool, error) {
	val, found := c.d[key]
	if !found {
		return false, nil
	}
	return val.(bool), nil
}
func (c *InMemoryCache) StoreString(key, value string) error {
	c.store(key, value)
	return nil
}
func (c *InMemoryCache) StoreExpiringString(key, value string, expiration time.Duration) error {
	c.store(key, value)
	return nil
}
func (c *InMemoryCache) FetchString(key string) (string, error) {
	val, found := c.d[key]
	if !found {
		return "", nil
	}
	return val.(string), nil
}

func (c *InMemoryCache) store(key string, value any) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.d[key] = value
}

// Package cachetest provides a fake cache for testing.
package cachetest

import (
	"sync"
	"time"
)

// NewInMemory creates a new InMemoryCache for test.
func NewInMemory() *InMemory {
	m := map[string]any{}
	c := InMemory{d: m}
	return &c
}

// InMemory implements a simple map-based cache for testing.
type InMemory struct {
	d   map[string]any
	mtx sync.Mutex
}

func (c *InMemory) StoreBool(key string, value bool) error {
	c.store(key, value)
	return nil
}
func (c *InMemory) StoreExpiringBool(key string, value bool, expiration time.Duration) error {
	c.store(key, value)
	return nil
}
func (c *InMemory) FetchBool(key string) (bool, error) {
	val, found := c.d[key]
	if !found {
		return false, nil
	}
	return val.(bool), nil
}
func (c *InMemory) StoreString(key, value string) error {
	c.store(key, value)
	return nil
}
func (c *InMemory) StoreExpiringString(key, value string, expiration time.Duration) error {
	c.store(key, value)
	return nil
}
func (c *InMemory) FetchString(key string) (string, error) {
	val, found := c.d[key]
	if !found {
		return "", nil
	}
	return val.(string), nil
}

func (c *InMemory) store(key string, value any) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.d[key] = value
}

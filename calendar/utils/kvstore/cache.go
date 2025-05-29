// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package kvstore

import (
	"sync"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
)

const (
	// Default TTL for cache entries if not specified
	DefaultCacheTTL = 5 * time.Minute
	// Default cleanup interval for expired entries
	DefaultCleanupInterval = 10 * time.Minute
)

// cacheEntry represents a single cache entry with its expiration time
type cacheEntry struct {
	Data       []byte
	Expiration time.Time
}

// CacheStore provides an in-memory key-value store with TTL (Time To Live) functionality
// that acts as a cache layer over another KVStore implementation
type cacheStore struct {
	mutex           sync.RWMutex
	entries         map[string]cacheEntry
	defaultTTL      time.Duration
	cleanupInterval time.Duration
	cleanupTimer    *time.Timer
	store           KVStore // Underlying persistent store
}

// Ensure cacheStore implements the KVStore interface
var _ KVStore = (*cacheStore)(nil)

// NewCacheStore creates a new in-memory cache store with TTL functionality
func NewCacheStore(underlying KVStore) KVStore {
	return NewCacheStoreWithOptions(underlying, DefaultCacheTTL, DefaultCleanupInterval)
}

// NewCacheStoreWithOptions creates a new in-memory cache store with custom TTL and cleanup interval
func NewCacheStoreWithOptions(underlying KVStore, defaultTTL, cleanupInterval time.Duration) KVStore {
	store := &cacheStore{
		entries:         make(map[string]cacheEntry),
		defaultTTL:      defaultTTL,
		cleanupInterval: cleanupInterval,
		store:           underlying,
	}

	// Start the cleanup timer
	store.startCleanupTimer()

	return store
}

// Load retrieves a value from the cache by key
// First checks memory cache, falls back to underlying store if not found
func (s *cacheStore) Load(key string) ([]byte, error) {
	s.mutex.RLock()
	entry, ok := s.entries[key]
	s.mutex.RUnlock()

	if ok {
		// Check if the entry has expired
		if time.Now().After(entry.Expiration) {
			// Lazy deletion on access
			go s.Delete(key)
		} else {
			// Return cached value that's not expired
			return entry.Data, nil
		}
	}

	// Not in cache or expired, try to load from underlying store
	data, err := s.store.Load(key)
	if err != nil {
		return nil, err
	}

	// Cache the value for future accesses
	s.mutex.Lock()
	defer s.mutex.Unlock()

	expiration := time.Now().Add(s.defaultTTL)
	s.entries[key] = cacheEntry{
		Data:       data,
		Expiration: expiration,
	}

	return data, nil
}

// Store adds or updates a value in the cache with the default TTL
// and also stores in the underlying persistent store
func (s *cacheStore) Store(key string, data []byte) error {
	// Store in underlying store first
	if err := s.store.Store(key, data); err != nil {
		return err
	}

	// Then update cache
	s.mutex.Lock()
	defer s.mutex.Unlock()

	expiration := time.Now().Add(s.defaultTTL)

	s.entries[key] = cacheEntry{
		Data:       data,
		Expiration: expiration,
	}

	return nil
}

// StoreTTL adds or updates a value in the cache with a specific TTL
// and also stores in the underlying persistent store with TTL if supported
func (s *cacheStore) StoreTTL(key string, data []byte, ttlSeconds int64) error {
	// Store in underlying store with TTL if possible
	if err := s.store.StoreTTL(key, data, ttlSeconds); err != nil {
		return err
	}

	// Then update cache
	s.mutex.Lock()
	defer s.mutex.Unlock()

	expiration := time.Now().Add(time.Duration(ttlSeconds) * time.Second)

	s.entries[key] = cacheEntry{
		Data:       data,
		Expiration: expiration,
	}

	return nil
}

// StoreWithOptions adds or updates a value in the cache with plugin KV set options
// and also stores in the underlying persistent store
func (s *cacheStore) StoreWithOptions(key string, value []byte, opts model.PluginKVSetOptions) (bool, error) {
	// Try store in underlying store first
	success, err := s.store.StoreWithOptions(key, value, opts)
	if err != nil || !success {
		return success, err
	}

	// If successful in underlying store, update the cache
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Calculate expiration based on TTL option
	expiration := time.Now().Add(s.defaultTTL)
	if opts.ExpireInSeconds > 0 {
		expiration = time.Now().Add(time.Duration(opts.ExpireInSeconds) * time.Second)
	}

	// Update the cache
	s.entries[key] = cacheEntry{
		Data:       value,
		Expiration: expiration,
	}

	return true, nil
}

// Delete removes a value from the cache by key
// and also deletes from the underlying persistent store
func (s *cacheStore) Delete(key string) error {
	// Delete from underlying store first
	if err := s.store.Delete(key); err != nil {
		return err
	}

	// Then remove from cache
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.entries, key)
	return nil
}

// List returns a paginated list of keys from the underlying store
func (s *cacheStore) List(page, perPage int) ([]string, error) {
	return s.store.List(page, perPage)
}

// startCleanupTimer starts a timer to periodically clean up expired entries
func (s *cacheStore) startCleanupTimer() {
	s.cleanupTimer = time.AfterFunc(s.cleanupInterval, func() {
		s.cleanup()
		s.startCleanupTimer()
	})
}

// cleanup removes all expired entries from the cache
func (s *cacheStore) cleanup() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	for key, entry := range s.entries {
		if now.After(entry.Expiration) {
			delete(s.entries, key)
		}
	}
}

// Stop stops the cleanup timer
func (s *cacheStore) Stop() {
	if s.cleanupTimer != nil {
		s.cleanupTimer.Stop()
	}
}

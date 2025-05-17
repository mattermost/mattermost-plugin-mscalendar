// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package kvstore

import (
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock KVStore for testing
type mockKVStore struct {
	mock.Mock
}

func (m *mockKVStore) Load(key string) ([]byte, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockKVStore) Store(key string, data []byte) error {
	args := m.Called(key, data)
	return args.Error(0)
}

func (m *mockKVStore) StoreTTL(key string, data []byte, ttlSeconds int64) error {
	args := m.Called(key, data, ttlSeconds)
	return args.Error(0)
}

func (m *mockKVStore) StoreWithOptions(key string, value []byte, opts model.PluginKVSetOptions) (bool, error) {
	args := m.Called(key, value, opts)
	return args.Bool(0), args.Error(1)
}

func (m *mockKVStore) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func TestCacheStore_LoadAndStore(t *testing.T) {
	mockStore := new(mockKVStore)
	testKey := "testkey"
	testData := []byte("testdata")

	// Set up mock expectations
	mockStore.On("Store", testKey, testData).Return(nil)
	mockStore.On("Load", testKey).Return(testData, nil)

	// Create cache store with mock
	store := NewCacheStore(mockStore)

	// Test storing data
	err := store.Store(testKey, testData)
	require.NoError(t, err)

	// Test loading data - should come from cache, not underlying store
	data, err := store.Load(testKey)
	require.NoError(t, err)
	assert.Equal(t, testData, data)

	// Verify underlying store was called only once for load (first miss)
	mockStore.AssertNumberOfCalls(t, "Load", 0) // Should be from cache, not underlying

	// Test loading non-existent key
	nonExistentKey := "nonexistent"
	mockStore.On("Load", nonExistentKey).Return(nil, ErrNotFound)
	_, err = store.Load(nonExistentKey)
	assert.Equal(t, ErrNotFound, err)
}

func TestCacheStore_LoadFromUnderlyingStore(t *testing.T) {
	mockStore := new(mockKVStore)
	testKey := "testkey"
	testData := []byte("testdata")

	// Set up mock to simulate data in underlying store but not in cache
	mockStore.On("Load", testKey).Return(testData, nil)

	// Create cache store with mock
	store := NewCacheStore(mockStore)

	// Data should be fetched from underlying store and cached
	data, err := store.Load(testKey)
	require.NoError(t, err)
	assert.Equal(t, testData, data)

	// Verify underlying store was called
	mockStore.AssertCalled(t, "Load", testKey)

	// Load again - should be from cache
	data, err = store.Load(testKey)
	require.NoError(t, err)
	assert.Equal(t, testData, data)

	// Verify underlying store was only called once
	mockStore.AssertNumberOfCalls(t, "Load", 1)
}

func TestCacheStore_Delete(t *testing.T) {
	mockStore := new(mockKVStore)
	testKey := "testkey"
	testData := []byte("testdata")

	// Set up mock expectations
	mockStore.On("Store", testKey, testData).Return(nil)
	mockStore.On("Delete", testKey).Return(nil)
	mockStore.On("Load", testKey).Return(nil, ErrNotFound)

	// Create cache store with mock
	store := NewCacheStore(mockStore)

	// Store data
	err := store.Store(testKey, testData)
	require.NoError(t, err)

	// Delete data
	err = store.Delete(testKey)
	require.NoError(t, err)

	// Data should no longer be available
	_, err = store.Load(testKey)
	assert.Equal(t, ErrNotFound, err)

	// Verify underlying store methods were called
	mockStore.AssertCalled(t, "Delete", testKey)
}

func TestCacheStore_TTL(t *testing.T) {
	mockStore := new(mockKVStore)
	testKey := "testkey"
	testData := []byte("testdata")
	ttl := int64(1) // 1 second TTL

	// Set up mock expectations
	mockStore.On("StoreTTL", testKey, testData, ttl).Return(nil)
	mockStore.On("Load", testKey).Return(testData, nil)

	// Create cache with short TTL for testing
	store := NewCacheStoreWithOptions(mockStore, 100*time.Millisecond, 1*time.Second)

	// Store data with short TTL
	err := store.StoreTTL(testKey, testData, ttl)
	require.NoError(t, err)

	// Data should be available immediately from cache
	data, err := store.Load(testKey)
	require.NoError(t, err)
	assert.Equal(t, testData, data)

	// Wait for cache TTL to expire
	time.Sleep(1100 * time.Millisecond)

	// Data should be retrieved from underlying store
	mockStore.On("Load", testKey).Return(testData, nil).Once()
	data, err = store.Load(testKey)
	require.NoError(t, err)
	assert.Equal(t, testData, data)

	// Verify underlying store load was called after cache expired
	mockStore.AssertNumberOfCalls(t, "Load", 1)
}

func TestCacheStore_StoreWithOptions(t *testing.T) {
	mockStore := new(mockKVStore)
	testKey := "testkey"
	initialData := []byte("initialdata")
	newData := []byte("newdata")

	// Create options for testing
	opts := model.PluginKVSetOptions{
		Atomic:   true,
		OldValue: initialData,
	}

	// Set up mock expectations
	mockStore.On("Store", testKey, initialData).Return(nil)
	mockStore.On("StoreWithOptions", testKey, newData, opts).Return(true, nil)

	// Create cache store with mock
	store := NewCacheStore(mockStore)

	// Store initial data
	err := store.Store(testKey, initialData)
	require.NoError(t, err)

	// Test atomic update with options
	success, err := store.StoreWithOptions(testKey, newData, opts)
	require.NoError(t, err)
	assert.True(t, success)

	// Verify data was updated in cache
	data, err := store.Load(testKey)
	require.NoError(t, err)
	assert.Equal(t, newData, data)

	// Verify underlying store method was called
	mockStore.AssertCalled(t, "StoreWithOptions", testKey, newData, opts)
}

func TestCacheStore_Cleanup(t *testing.T) {
	mockStore := new(mockKVStore)

	// Set up mock expectations for each key
	for i := 0; i < 5; i++ {
		key := "key" + string(rune('a'+i))
		mockStore.On("StoreTTL", key, []byte("data"), int64(1)).Return(nil)
	}

	// Create cache with manual cleanup for testing
	store := &cacheStore{
		entries:         make(map[string]cacheEntry),
		defaultTTL:      50 * time.Millisecond,
		cleanupInterval: 100 * time.Millisecond,
		store:           mockStore,
	}

	// Add several entries with very short TTL
	for i := 0; i < 5; i++ {
		key := "key" + string(rune('a'+i))
		// Store with extremely short TTL for testing
		store.StoreTTL(key, []byte("data"), 1) // 1 second TTL
	}

	// Initially, all entries should be in the cache
	store.mutex.RLock()
	initialCount := len(store.entries)
	store.mutex.RUnlock()
	assert.Equal(t, 5, initialCount)

	// Manually force an immediate cleanup
	time.Sleep(1100 * time.Millisecond) // Wait for TTL to expire
	store.cleanup()                     // Force cleanup immediately

	// After cleanup, cache should be empty
	store.mutex.RLock()
	finalCount := len(store.entries)
	store.mutex.RUnlock()
	assert.Equal(t, 0, finalCount)
}

func TestCacheStore_Prefixing(t *testing.T) {
	mockStore := new(mockKVStore)
	testKey := "key1"
	testData := []byte("data1")

	// Set up mock expectations
	mockStore.On("Store", testKey, testData).Return(nil)

	// Create cache store with mock
	store := NewCacheStore(mockStore).(*cacheStore)

	// Store data
	err := store.Store(testKey, testData)
	require.NoError(t, err)

	// Check that the key was stored
	store.mutex.RLock()
	_, exists := store.entries[testKey]
	store.mutex.RUnlock()
	assert.True(t, exists)

	// Verify underlying store method was called
	mockStore.AssertCalled(t, "Store", testKey, testData)
}

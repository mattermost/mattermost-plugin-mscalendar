// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package kvstore

import (
	"time"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/jellydator/ttlcache/v3"
)

type cacheKeyStore struct {
	store KVStore
	ttl   time.Duration
	cache *ttlcache.Cache[string, []byte]
}

var _ KVStore = (*cacheKeyStore)(nil)

func NewCacheKeyStore(s KVStore, ttl time.Duration) KVStore {
	cache := ttlcache.New[string, []byte](
		ttlcache.WithTTL[string, []byte](ttl),
	)
	go cache.Start() // Expiration job

	return &cacheKeyStore{
		store: s,
		cache: cache,
	}
}

func (s cacheKeyStore) storeCache(key string, value []byte) error {
	s.cache.Set(key, value, s.ttl)
	return nil
}

func (s cacheKeyStore) loadCache(key string) ([]byte, bool) {
	item := s.cache.Get(key)
	if item == nil {
		return nil, false
	}

	return item.Value(), true
}

func (s cacheKeyStore) Load(key string) ([]byte, error) {
	if value, exists := s.loadCache(key); exists {
		return value, nil
	}

	value, err := s.store.Load(key)
	if err == nil {
		s.storeCache(key, value)
	}

	return value, err
}

func (s cacheKeyStore) Store(key string, value []byte) error {
	s.storeCache(key, value)

	return s.store.Store(key, value)
}

func (s cacheKeyStore) StoreTTL(key string, value []byte, ttlSeconds int64) error {
	s.storeCache(key, value)

	return s.store.StoreTTL(key, value, ttlSeconds)
}

func (s cacheKeyStore) StoreWithOptions(key string, value []byte, opts model.PluginKVSetOptions) (bool, error) {
	s.storeCache(key, value)
	return s.store.StoreWithOptions(key, value, opts)
}

func (s cacheKeyStore) Delete(key string) error {
	s.cache.Delete(key)
	return s.store.Delete(key)
}

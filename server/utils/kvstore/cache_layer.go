package kvstore

import (
	"encoding/json"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/services/cache"
	"github.com/mattermost/mattermost-server/v5/services/cache/lru"
)

const (
	DirtyListTTL = 10 * time.Second
)

type CacheLayer struct {
	store          KVStore
	lastDirtyCheck time.Time
	cache          cache.Cache
	key            string
	dirtyList      []string
}

func NewCachedStore(s KVStore, cacheKey string, cacheSize int) KVStore {
	return &CacheLayer{
		store:     s,
		cache:     lru.New(cacheSize),
		key:       cacheKey,
		dirtyList: []string{},
	}
}

func (s *CacheLayer) Load(key string) ([]byte, error) {
	dirty := s.checkDirty(key)

	var cacheItem interface{}
	var cacheFound bool
	if !dirty {
		cacheItem, cacheFound = s.cache.Get(key)
	}
	if dirty || !cacheFound {
		data, err := s.store.Load(key)
		if err != nil {
			s.cache.Add(key, data)
		}
		return data, err
	}

	return cacheItem.([]byte), nil
}

func (s *CacheLayer) Store(key string, data []byte) error {
	err := s.store.Store(key, data)
	if err != nil {
		return err
	}
	s.cache.Add(key, data)
	err = s.addDirty(key)
	if err != nil {
		return err
	}

	return nil
}
func (s *CacheLayer) StoreTTL(key string, data []byte, ttlSeconds int64) error {
	err := s.store.StoreTTL(key, data, ttlSeconds)
	if err != nil {
		return err
	}
	s.cache.AddWithExpiresInSecs(key, data, ttlSeconds)
	err = s.addDirtyTTL(key, ttlSeconds)
	if err != nil {
		return err
	}

	return nil
}
func (s *CacheLayer) StoreWithOptions(key string, value []byte, opts model.PluginKVSetOptions) (bool, error) {
	ok, err := s.store.StoreWithOptions(key, value, opts)
	if err != nil {
		return ok, err
	}
	if !ok {
		return ok, err
	}

	s.cache.Add(key, value)
	err = s.addDirty(key)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (s *CacheLayer) Delete(key string) error {
	err := s.store.Delete(key)
	if err != nil {
		return err
	}

	s.cache.Remove(key)
	err = s.addDirty(key)
	if err != nil {
		return err
	}

	return nil
}

func (s *CacheLayer) CompareAndDelete(key string, oldValue []byte) (bool, error) {
	return s.store.CompareAndDelete(key, oldValue)
}

func (s *CacheLayer) ClearCaches() {
	s.cache.Purge()
}

func (s *CacheLayer) addDirty(key string) error {
	return AtomicModify(s, s.key, func(initialValue []byte, storeErr error) ([]byte, error) {
		if storeErr != nil && storeErr != ErrNotFound {
			return nil, storeErr
		}

		if initialValue == nil {
			initialValue = make([]byte, 0, 1)
		}

		storedList := []string{}
		if len(initialValue) > 0 {
			err := json.Unmarshal(initialValue, &storedList)
			if err != nil {
				return nil, err
			}
		}

		for _, listKey := range storedList {
			if listKey == key {
				return initialValue, nil
			}
		}

		updated := append(storedList, key)

		b, err := json.Marshal(updated)
		if err != nil {
			return nil, err
		}

		return b, nil
	})
}

func (s *CacheLayer) addDirtyTTL(key string, ttl int64) error {
	return s.addDirty(key)
}

func (s *CacheLayer) checkDirty(key string) bool {
	if s.lastDirtyCheck.Add(DirtyListTTL).Before(time.Now()) {
		err := AtomicDelete(s.store, s.key, func(initialValue []byte, storeErr error) error {
			if storeErr != nil && storeErr != ErrNotFound {
				return storeErr
			}

			storedList := []string{}
			if len(initialValue) > 0 {
				err := json.Unmarshal(initialValue, &storedList)
				if err != nil {
					return err
				}

				s.mergeDirtyList(storedList)
			}

			return nil
		})
		if err != nil {
			return true
		}
	}

	for i, dirtyKey := range s.dirtyList {
		if dirtyKey == key {
			s.dirtyList = append(s.dirtyList[:i], s.dirtyList[i+1:]...)
			return true
		}
	}

	return false
}

func (s *CacheLayer) mergeDirtyList(newList []string) {
Loop:
	for _, key := range newList {
		for _, storedKey := range s.dirtyList {
			if key == storedKey {
				continue Loop
			}
		}

		s.dirtyList = append(s.dirtyList, key)
	}
}

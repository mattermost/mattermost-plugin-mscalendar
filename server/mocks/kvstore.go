package mocks

import (
	"errors"

	"github.com/mattermost/mattermost-plugin-msoffice/server/kvstore"
)

var _ kvstore.KVStore = &MockKVStore{}

type MockKVStore struct {
	KVStore map[string][]byte
	Err     error
}

func (kv *MockKVStore) Load(key string) ([]byte, error) {
	empty := make([]byte, 0)

	if kv.Err != nil {
		return empty, kv.Err
	}

	if value, exists := kv.KVStore[key]; exists {
		return value, nil
	}

	return empty, errors.New("key not found in store")
}

func (kv *MockKVStore) Store(key string, data []byte) error {
	if kv.Err != nil {
		return kv.Err
	}

	kv.KVStore[key] = data

	return nil
}

func (kv *MockKVStore) Delete(key string) error {
	if _, exists := kv.KVStore[key]; exists {
		delete(kv.KVStore, key)
	}

	return errors.New("key not found in store")
}

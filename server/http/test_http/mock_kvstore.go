package testhttp

import (
	"github.com/mattermost/mattermost-plugin-msoffice/server/kvstore"
	"github.com/stretchr/testify/mock"
)

var (
	_ kvstore.KVStore = &mockKVStore{}
)

func newMockKVStore(b []byte, err error) *mockKVStore {
	s := &mockKVStore{}

	s.On("Load", mock.Anything).Return(b, err)
	s.On("Store", mock.Anything, mock.Anything).Return(err)
	s.On("Delete", mock.Anything).Return(err)

	return s
}

type mockKVStore struct {
	mock.Mock
}

func (kv *mockKVStore) Load(key string) ([]byte, error) {
	args := kv.Called(key)
	return []byte(args.String(0)), args.Error(1)
}

func (kv *mockKVStore) Store(key string, data []byte) error {
	args := kv.Called(key, data)
	return args.Error(0)
}

func (kv *mockKVStore) Delete(key string) error {
	args := kv.Called(key)
	return args.Error(0)
}

package kvstore

import (
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockPluginAPI struct {
	plugin.API
	mock.Mock
}

func (m *MockPluginAPI) KVGet(key string) ([]byte, *model.AppError) {
	args := m.Called(key)
	data, _ := args.Get(0).([]byte)
	if err := args.Get(1); err != nil {
		return nil, err.(*model.AppError)
	}
	return data, nil
}

func (m *MockPluginAPI) KVSet(key string, data []byte) *model.AppError {
	args := m.Called(key, data)
	if err := args.Get(0); err != nil {
		return err.(*model.AppError)
	}
	return nil
}

func (m *MockPluginAPI) KVSetWithExpiry(key string, data []byte, ttl int64) *model.AppError {
	args := m.Called(key, data, ttl)
	if err := args.Get(0); err != nil {
		return err.(*model.AppError)
	}
	return nil
}

func (m *MockPluginAPI) KVSetWithOptions(key string, value []byte, options model.PluginKVSetOptions) (bool, *model.AppError) {
	args := m.Called(key, value, options)
	success := args.Bool(0)
	if err := args.Get(1); err != nil {
		return success, err.(*model.AppError)
	}
	return success, nil
}

func (m *MockPluginAPI) KVDelete(key string) *model.AppError {
	args := m.Called(key)
	if err := args.Get(0); err != nil {
		return err.(*model.AppError)
	}
	return nil
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		setup      func(*MockPluginAPI)
		assertions func(*testing.T, []byte, error)
	}{
		{
			name: "Error during KVGet",
			key:  "error-key",
			setup: func(mockAPI *MockPluginAPI) {
				mockAPI.On("KVGet", "error-key").Return(nil, &model.AppError{Message: "KVGet failed"})
			},
			assertions: func(t *testing.T, data []byte, err error) {
				require.Nil(t, data, "expected nil data")
				require.EqualError(t, err, "failed plugin KVGet: KVGet failed", "unexpected error message")
			},
		},
		{
			name: "Key not found",
			key:  "missing-key",
			setup: func(mockAPI *MockPluginAPI) {
				mockAPI.On("KVGet", "missing-key").Return(nil, nil)
			},
			assertions: func(t *testing.T, data []byte, err error) {
				require.Nil(t, data, "expected nil data")
				require.EqualError(t, err, ErrNotFound.Error(), "unexpected error message")
			},
		},
		{
			name: "Load successfully",
			key:  "test-key",
			setup: func(mockAPI *MockPluginAPI) {
				mockAPI.On("KVGet", "test-key").Return([]byte("test-value"), nil)
			},
			assertions: func(t *testing.T, data []byte, err error) {
				require.Equal(t, []byte("test-value"), data, "unexpected data returned")
				require.NoError(t, err, "unexpected error occurred")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &MockPluginAPI{}
			store := NewPluginStore(mockAPI)
			tt.setup(mockAPI)

			data, err := store.Load(tt.key)

			tt.assertions(t, data, err)

			mockAPI.AssertExpectations(t)
		})
	}
}

func TestStore(t *testing.T) {
	tests := []struct {
		name       string
		expiryTime int
		setup      func(*MockPluginAPI)
		assertions func(*testing.T, error)
	}{
		{
			name:       "Error during KVSet with TTL",
			expiryTime: 60,
			setup: func(mockAPI *MockPluginAPI) {
				mockAPI.On("KVSetWithExpiry", "mockKey", []byte("mockValue"), int64(60)).Return(&model.AppError{Message: "KVSet failed"})
			},
			assertions: func(t *testing.T, err error) {
				require.EqualError(t, err, "failed plugin KVSet (ttl: 60s) \"mockKey\": KVSet failed", "unexpected error message")
			},
		},
		{
			name:       "Error during KVSet without TTL",
			expiryTime: 0,
			setup: func(mockAPI *MockPluginAPI) {
				mockAPI.On("KVSet", "mockKey", []byte("mockValue")).Return(&model.AppError{Message: "KVSet failed"})
			},
			assertions: func(t *testing.T, err error) {
				require.EqualError(t, err, "failed plugin KVSet (ttl: 0s) \"mockKey\": KVSet failed", "unexpected error message")
			},
		},
		{
			name:       "Store with TTL successfully",
			expiryTime: 60,
			setup: func(mockAPI *MockPluginAPI) {
				mockAPI.On("KVSetWithExpiry", "mockKey", []byte("mockValue"), int64(60)).Return(nil)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err, "unexpected error occurred")
			},
		},
		{
			name:       "Store without TTL successfully",
			expiryTime: 0,
			setup: func(mockAPI *MockPluginAPI) {
				mockAPI.On("KVSet", "mockKey", []byte("mockValue")).Return(nil)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err, "unexpected error occurred")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &MockPluginAPI{}
			store := NewPluginStoreWithExpiry(mockAPI, time.Duration(tt.expiryTime)*time.Second)
			tt.setup(mockAPI)

			err := store.Store("mockKey", []byte("mockValue"))

			tt.assertions(t, err)

			mockAPI.AssertExpectations(t)
		})
	}
}

func TestStoreTTL(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*MockPluginAPI)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error during storing with TTL",
			setup: func(mockAPI *MockPluginAPI) {
				mockAPI.On("KVSetWithExpiry", "mockKey", []byte("mockValue"), int64(60)).Return(&model.AppError{Message: "KVSet failed"})
			},
			assertions: func(t *testing.T, err error) {
				require.EqualError(t, err, "failed plugin KVSet (ttl: 60s) \"mockKey\": KVSet failed", "unexpected error message")
			},
		},
		{
			name: "Store with TTL successfully",
			setup: func(mockAPI *MockPluginAPI) {
				mockAPI.On("KVSetWithExpiry", "mockKey", []byte("mockValue"), int64(60)).Return(nil)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err, "unexpected error occurred")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &MockPluginAPI{}
			store := NewPluginStoreWithExpiry(mockAPI, 60*time.Second)
			tt.setup(mockAPI)

			err := store.StoreTTL("mockKey", []byte("mockValue"), 60)

			tt.assertions(t, err)

			mockAPI.AssertExpectations(t)
		})
	}
}

func TestStoreWithOptions(t *testing.T) {
	tests := []struct {
		name       string
		expiryTime int64
		opts       model.PluginKVSetOptions
		setup      func(*MockPluginAPI)
		assertions func(*testing.T, bool, error)
	}{
		{
			name:       "Error during KVSetWithOptions",
			expiryTime: 60,
			opts: model.PluginKVSetOptions{
				ExpireInSeconds: 30,
			},
			setup: func(mockAPI *MockPluginAPI) {
				mockAPI.On("KVSetWithOptions", "mockKey", []byte("mockValue"), model.PluginKVSetOptions{ExpireInSeconds: 30}).Return(false, &model.AppError{Message: "KVSet failed"})
			},
			assertions: func(t *testing.T, success bool, err error) {
				require.False(t, success, "expected success to be false")
				require.EqualError(t, err, "failed plugin KVSet (ttl: 30s) \"mockKey\": KVSet failed", "unexpected error message")
			},
		},
		{
			name:       "Use default TTL when opts.ExpireInSeconds is 0",
			expiryTime: 60,
			opts:       model.PluginKVSetOptions{},
			setup: func(mockAPI *MockPluginAPI) {
				mockAPI.On("KVSetWithOptions", "mockKey", []byte("mockValue"), model.PluginKVSetOptions{ExpireInSeconds: 60}).Return(true, nil)
			},
			assertions: func(t *testing.T, success bool, err error) {
				require.True(t, success, "expected success to be true")
				require.NoError(t, err, "unexpected error occurred")
			},
		},
		{
			name:       "Store with options successfully",
			expiryTime: 60,
			opts: model.PluginKVSetOptions{
				ExpireInSeconds: 30,
			},
			setup: func(mockAPI *MockPluginAPI) {
				mockAPI.On("KVSetWithOptions", "mockKey", []byte("mockValue"), model.PluginKVSetOptions{ExpireInSeconds: 30}).Return(true, nil)
			},
			assertions: func(t *testing.T, success bool, err error) {
				require.True(t, success, "expected success to be true")
				require.NoError(t, err, "unexpected error occurred")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &MockPluginAPI{}
			store := NewPluginStoreWithExpiry(mockAPI, time.Duration(tt.expiryTime)*time.Second)
			tt.setup(mockAPI)

			success, err := store.StoreWithOptions("mockKey", []byte("mockValue"), tt.opts)

			tt.assertions(t, success, err)

			mockAPI.AssertExpectations(t)
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*MockPluginAPI)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error during KVDelete",
			setup: func(mockAPI *MockPluginAPI) {
				mockAPI.On("KVDelete", "mockKey").Return(&model.AppError{Message: "KVDelete failed"})
			},
			assertions: func(t *testing.T, err error) {
				require.EqualError(t, err, "failed plugin KVdelete \"mockKey\": KVDelete failed", "unexpected error message")
			},
		},
		{
			name: "Delete successfully",
			setup: func(mockAPI *MockPluginAPI) {
				mockAPI.On("KVDelete", "mockKey").Return(nil)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err, "unexpected error occurred")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &MockPluginAPI{}
			store := NewPluginStore(mockAPI)
			tt.setup(mockAPI)

			err := store.Delete("mockKey")

			tt.assertions(t, err)

			mockAPI.AssertExpectations(t)
		})
	}
}

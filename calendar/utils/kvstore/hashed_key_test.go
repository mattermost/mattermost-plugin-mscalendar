// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package kvstore

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func Test_hashKey(t *testing.T) {
	type args struct {
		prefix      string
		hashableKey string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{"", ""}, ""},
		{"value", args{"", "https://mmtest.mattermost.com"}, "53d1d6fa60f26d84e2087f61d535d073"},
		{"prefix", args{"abc_", ""}, "abc_"},
		{"prefix value", args{"abc_", "123"}, "abc_202cb962ac59075b964b07152d234b70"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hashKey(tt.args.prefix, tt.args.hashableKey)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestHashedKeyStoreList(t *testing.T) {
	tests := []struct {
		name          string
		storePrefix   string
		mockKeys      []string
		mockErr       error
		expectedKeys  []string
		expectedError string
	}{
		{
			name:         "List with no items",
			storePrefix:  "test_",
			mockKeys:     []string{},
			expectedKeys: []string{},
		},
		{
			name:         "List all items",
			storePrefix:  "test_",
			mockKeys:     []string{"test_key1", "test_key2"},
			expectedKeys: []string{"test_key1", "test_key2"},
		},
		{
			name:         "List with prefix filter",
			storePrefix:  "test_",
			mockKeys:     []string{"test_key1", "test_202cb962ac59075b964b07152d234b70"},
			expectedKeys: []string{"test_key1", "test_202cb962ac59075b964b07152d234b70"},
		},
		{
			name:          "List with error",
			storePrefix:   "test_",
			mockErr:       errors.New("mock error"),
			expectedError: "mock error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(mockKVStore)

			// Set expectation on the List method
			mockStore.On("List", 0, 100).Return(tt.mockKeys, tt.mockErr)

			store := NewHashedKeyStore(mockStore, tt.storePrefix)

			keys, err := store.List(0, 100)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedKeys, keys)
			}

			mockStore.AssertExpectations(t)
		})
	}
}

// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package kvstore

import (
	"time"

	"github.com/mattermost/mattermost-server/plugin"
)

// OneTimeStore is a KV store that deletes each record after the first load,
type OneTimeStore KVStore

type ots struct {
	KVStore
}

func NewOneTimePluginStore(api plugin.API, ttl time.Duration) OneTimeStore {
	return &ots{
		KVStore: NewPluginStoreWithExpiry(api, ttl),
	}
}

func NewOneTimeStore(kv KVStore) OneTimeStore {
	return &ots{
		KVStore: kv,
	}
}

func (s ots) Load(key string) (data []byte, returnErr error) {
	data, err := s.KVStore.Load(key)

	_ = s.Delete(key)

	return data, err
}

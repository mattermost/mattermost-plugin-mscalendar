// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"time"

	"github.com/mattermost/mattermost-server/plugin"

	"github.com/mattermost/mattermost-plugin-msoffice/server/kvstore"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
)

const (
	UserKeyPrefix             = "user_"
	MattermostUserIDKeyPrefix = "mmuid_"
	OAuth2KeyPrefix           = "oauth2_"
	SubscriptionKeyPrefix     = "sub_"
)

const OAuth2KeyExpiration = 15 * time.Minute

type Store interface {
	UserStore
	OAuth2StateStore
	SubscriptionStore
}

type pluginStore struct {
	basicKV            kvstore.KVStore
	oauth2KV           kvstore.KVStore
	userKV             kvstore.KVStore
	mattermostUserIDKV kvstore.KVStore
	subscriptionKV     kvstore.KVStore
	Logger             utils.Logger
}

func NewPluginStore(api plugin.API) Store {
	basicKV := kvstore.NewPluginStore(api)
	return &pluginStore{
		basicKV:            basicKV,
		userKV:             kvstore.NewHashedKeyStore(basicKV, UserKeyPrefix),
		mattermostUserIDKV: kvstore.NewHashedKeyStore(basicKV, MattermostUserIDKeyPrefix),
		subscriptionKV:     kvstore.NewHashedKeyStore(basicKV, SubscriptionKeyPrefix),
		oauth2KV:           kvstore.NewHashedKeyStore(kvstore.NewOneTimePluginStore(api, OAuth2KeyExpiration), OAuth2KeyPrefix),
		Logger:             api,
	}
}

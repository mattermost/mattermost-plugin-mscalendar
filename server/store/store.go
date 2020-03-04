// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"time"

	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/kvstore"
)

const (
	UserKeyPrefix             = "user_"
	UserIndexKeyPrefix        = "userindex_"
	MattermostUserIDKeyPrefix = "mmuid_"
	OAuth2KeyPrefix           = "oauth2_"
	SubscriptionKeyPrefix     = "sub_"
	EventKeyPrefix            = "ev_"
	WelcomeKeyPrefix          = "welcome_"
)

const OAuth2KeyExpiration = 15 * time.Minute

var ErrNotFound = kvstore.ErrNotFound

type Store interface {
	UserStore
	OAuth2StateStore
	SubscriptionStore
	EventStore
	WelcomeStore
}

type pluginStore struct {
	basicKV            kvstore.KVStore
	oauth2KV           kvstore.KVStore
	userKV             kvstore.KVStore
	mattermostUserIDKV kvstore.KVStore
	userIndexKV        kvstore.KVStore
	subscriptionKV     kvstore.KVStore
	eventKV            kvstore.KVStore
	welcomeIndexKV     kvstore.KVStore
	Logger             bot.Logger
}

func NewPluginStore(api plugin.API, logger bot.Logger) Store {
	basicKV := kvstore.NewPluginStore(api)
	return &pluginStore{
		basicKV:            basicKV,
		userKV:             kvstore.NewHashedKeyStore(basicKV, UserKeyPrefix),
		userIndexKV:        kvstore.NewHashedKeyStore(basicKV, UserIndexKeyPrefix),
		mattermostUserIDKV: kvstore.NewHashedKeyStore(basicKV, MattermostUserIDKeyPrefix),
		subscriptionKV:     kvstore.NewHashedKeyStore(basicKV, SubscriptionKeyPrefix),
		eventKV:            kvstore.NewHashedKeyStore(basicKV, EventKeyPrefix),
		oauth2KV:           kvstore.NewHashedKeyStore(kvstore.NewOneTimePluginStore(api, OAuth2KeyExpiration), OAuth2KeyPrefix),
		welcomeIndexKV:     kvstore.NewHashedKeyStore(basicKV, WelcomeKeyPrefix),
		Logger:             logger,
	}
}

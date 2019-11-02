// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package user

import (
	"errors"
	"time"

	"github.com/mattermost/mattermost-server/plugin"

	"github.com/mattermost/mattermost-plugin-msoffice/server/kvstore"
)

const OAuth2KeyPrefix = "user_"
const OAuth2KeyExpiration = 15 * time.Minute

type OAuth2StateStore interface {
	Verify(state string) error
	Store(state string) error
}

type oauth2StateStore struct {
	kv kvstore.KVStore
}

func NewOAuth2StateStore(api plugin.API) OAuth2StateStore {
	return &oauth2StateStore{
		kv: kvstore.NewHashedKeyStore(kvstore.NewOneTimePluginStore(api, OAuth2KeyExpiration), OAuth2KeyPrefix),
	}
}

func (s *oauth2StateStore) Verify(state string) error {
	data, err := s.kv.Load(state)
	if err != nil {
		return err
	}
	if string(data) != state {
		return errors.New("authentication attempt expired, please try again.")
	}
	return nil
}

func (s *oauth2StateStore) Store(state string) error {
	return s.kv.Store(state, []byte(state))
}

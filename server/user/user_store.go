// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package user

import (
	"github.com/mattermost/mattermost-plugin-msoffice/server/kvstore"
)

const UserKeyPrefix = "user_"

type Store interface {
	LoadRemoteUser(mattermostUserId string, ref interface{}) error
	LoadMattermostUserId(remoteUserId string) (string, error)
	Store(mattermostUserId string, user *User) error
	Delete(mattermostUserId string) error
}

type store struct {
	kv kvstore.KVStore
}

func NewStore(s kvstore.KVStore) Store {
	return &store{
		kv: kvstore.NewHashedKeyStore(s, UserKeyPrefix),
	}
}

func (s *store) LoadRemoteUser(mattermostUserId string, ref interface{}) error {
	return kvstore.LoadJSON(s.kv, mattermostUserId, ref)
}

func (s *store) LoadMattermostUserId(remoteUserId string) (string, error) {
	data, err := s.kv.Load(remoteUserId)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *store) Store(mattermostUserId string, user *User) error {
	err := kvstore.StoreJSON(s.kv, mattermostUserId, user)
	if err != nil {
		return err
	}

	err = s.kv.Store(user.Remote.ID, []byte(mattermostUserId))
	if err != nil {
		_ = s.kv.Delete(mattermostUserId)
		return err
	}

	return nil
}

func (s *store) Delete(mattermostUserId string) error {
	u := User{}
	err := s.LoadRemoteUser(mattermostUserId, &u)
	if err != nil {
		return err
	}
	err = s.kv.Delete(mattermostUserId)
	if err != nil {
		return err
	}
	err = s.kv.Delete(u.Remote.ID)
	if err != nil {
		return err
	}
	return nil
}

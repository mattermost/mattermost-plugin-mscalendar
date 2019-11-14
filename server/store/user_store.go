// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"github.com/mattermost/mattermost-plugin-msoffice/server/kvstore"
)

type UserStore interface {
	LoadUser(mattermostUserId string) (*User, error)
	LoadMattermostUserId(remoteUserId string) (string, error)
	StoreUser(user *User) error
	DeleteUser(mattermostUserId string) error
}

func (s *pluginStore) LoadUser(mattermostUserId string) (*User, error) {
	user := User{}
	err := kvstore.LoadJSON(s.userKV, mattermostUserId, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *pluginStore) LoadMattermostUserId(remoteUserId string) (string, error) {
	data, err := s.userKV.Load(remoteUserId)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *pluginStore) StoreUser(user *User) error {
	err := kvstore.StoreJSON(s.userKV, user.MattermostUserID, user)
	if err != nil {
		return err
	}

	err = s.mattermostUserIDKV.Store(user.Remote.ID, []byte(user.MattermostUserID))
	if err != nil {
		_ = s.userKV.Delete(user.MattermostUserID)
		return err
	}

	return nil
}

func (s *pluginStore) DeleteUser(mattermostUserID string) error {
	u, err := s.LoadUser(mattermostUserID)
	if err != nil {
		return err
	}
	err = s.userKV.Delete(mattermostUserID)
	if err != nil {
		return err
	}
	err = s.mattermostUserIDKV.Delete(u.Remote.ID)
	if err != nil {
		return err
	}
	return nil
}

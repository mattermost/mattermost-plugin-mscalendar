// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"errors"
)

type OAuth2StateStore interface {
	VerifyOAuth2State(state string) error
	StoreOAuth2State(state string) error
}

func (s *pluginStore) VerifyOAuth2State(state string) error {
	data, err := s.subscriptionKV.Load(state)
	if err != nil {
		return err
	}
	if string(data) != state {
		return errors.New("authentication attempt expired, please try again.")
	}
	return nil
}

func (s *pluginStore) StoreOAuth2State(state string) error {
	return s.subscriptionKV.Store(state, []byte(state))
}

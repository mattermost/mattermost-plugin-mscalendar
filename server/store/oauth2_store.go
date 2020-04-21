// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

const (
	oAuth2StateTimeToLive = 300 // seconds
)

// OAuth2StateStore manages OAuth2 state
type OAuth2StateStore interface {
	VerifyOAuth2State(state string) error
	StoreOAuth2State(state string) error
}

func (s *pluginStore) VerifyOAuth2State(state string) error {
	data, err := s.oauth2KV.Load(state)
	if err != nil {
		return err
	}

	if string(data) != state || err == store.ErrNotFound {
		return errors.New("authentication attempt expired, please try again")
	}
	return nil
}

func (s *pluginStore) StoreOAuth2State(state string) error {
	return s.oauth2KV.StoreTTL(state, []byte(state), oAuth2StateTimeToLive)
}

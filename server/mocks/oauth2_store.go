package mocks

import (
	"github.com/mattermost/mattermost-plugin-msoffice/server/user"
)

var (
	_ user.OAuth2StateStore = &MockOAuth2StateStore{}
)

type MockOAuth2StateStore struct {
	Err error
}

func (s *MockOAuth2StateStore) Verify(state string) error {
	if s.Err != nil {
		return s.Err
	}

	return nil
}

func (s *MockOAuth2StateStore) Store(state string) error {
	if s.Err != nil {
		return s.Err
	}

	return nil
}

package testhttp

import (
	"github.com/mattermost/mattermost-plugin-msoffice/server/user"
	"github.com/stretchr/testify/mock"
)

var (
	_ user.OAuth2StateStore = &mockOAuth2StateStore{}
)

func newMockOAuth2StateStore(err error) *mockOAuth2StateStore {
	s := &mockOAuth2StateStore{}

	s.On("Verify", mock.Anything).Return(err)
	s.On("Store", mock.Anything).Return(err)

	return s
}

type mockOAuth2StateStore struct {
	mock.Mock
}

func (s *mockOAuth2StateStore) Verify(state string) error {
	args := s.Called(state)
	return args.Error(0)
}

func (s *mockOAuth2StateStore) Store(state string) error {
	args := s.Called(state)
	return args.Error(0)
}

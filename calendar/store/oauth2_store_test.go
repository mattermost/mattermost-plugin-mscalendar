package store

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/testutil"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestVerifyOAuth2State(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error loading state",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "oauth2_fb89cea34670836627b56ad5b94ce5e3").Return(nil, &model.AppError{Message: "Error getting state"}).Times(1)
				mockAPI.On("KVDelete", "oauth2_fb89cea34670836627b56ad5b94ce5e3").Return(nil)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.EqualError(t, err, "failed plugin KVGet: Error getting state")
			},
		},
		{
			name: "Invalid Oauth state",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "oauth2_fb89cea34670836627b56ad5b94ce5e3").Return([]byte("invalidState"), nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.EqualError(t, err, "invalid oauth state, please try again")
			},
		},
		{
			name: "Successfull Oauth state verification",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "oauth2_fb89cea34670836627b56ad5b94ce5e3").Return([]byte(MockState), nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Nil(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI, store, _, _, _ := GetMockSetup(t)
			tt.setup(mockAPI)

			err := store.VerifyOAuth2State(MockState)

			tt.assertions(t, err)

			mockAPI.AssertExpectations(t)
		})
	}
}

func TestStoreOAuth2State(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error loading state",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVSetWithExpiry", "oauth2_fb89cea34670836627b56ad5b94ce5e3", mock.Anything, int64(oAuth2StateTimeToLive)).Return(&model.AppError{Message: "Error loading state"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(t, err, "Error loading state")
			},
		},
		{
			name: "Successfull Oauth state verification",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVSetWithExpiry", "oauth2_fb89cea34670836627b56ad5b94ce5e3", mock.Anything, int64(oAuth2StateTimeToLive)).Return(nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Nil(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI, store, _, _, _ := GetMockSetup(t)
			tt.setup(mockAPI)

			err := store.StoreOAuth2State(MockState)

			tt.assertions(t, err)

			mockAPI.AssertExpectations(t)
		})
	}
}

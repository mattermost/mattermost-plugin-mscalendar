package store

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/testutil"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot/mock_bot"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestLoadSubscription(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, *Subscription, error)
	}{
		{
			name: "Error loading subscription",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "sub_0c47c5b7e2a88ec9256c8ac0e71b0f6e").Return(nil, &model.AppError{Message: "Subscription not found"}).Times(1)
			},
			assertions: func(t *testing.T, sub *Subscription, err error) {
				require.Error(t, err)
				require.Nil(t, sub)
				require.EqualError(t, err, "failed plugin KVGet: Subscription not found")
			},
		},
		{
			name: "Subscription successfully loaded",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "sub_0c47c5b7e2a88ec9256c8ac0e71b0f6e").Return([]byte(fmt.Sprintf(`{"PluginVersion":"1.0","Remote":{"ID":"%s","CreatorID":"%s"}}`, MockRemoteUserID, MockCreatorID)), nil).Times(1)
			},
			assertions: func(t *testing.T, sub *Subscription, err error) {
				require.NoError(t, err)
				require.NotNil(t, sub)
				require.Equal(t, "1.0", sub.PluginVersion)
				require.Equal(t, MockRemoteUserID, sub.Remote.ID)
				require.Equal(t, MockCreatorID, sub.Remote.CreatorID)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI, store, _, _, _ := GetMockSetup(t)
			tt.setup(mockAPI)

			sub, err := store.LoadSubscription(MockSubscriptionID)

			tt.assertions(t, sub, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestStoreUserSubscription(t *testing.T) {
	mockUser := GetMockUser()
	mockSubscription := GetMockSubscription()

	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI, *mock_bot.MockLogger, *mock_bot.MockLogger)
		assertions func(*testing.T, error)
	}{
		{
			name:  "User does not match subscription creator",
			setup: func(_ *testutil.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger) {},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.EqualError(t, err, `user "mockRemoteID" does not match the subscription creator "mockCreatorID"`)
			},
		},
		{
			name: "Error storing subscription",
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger) {
				mockSubscription.Remote.CreatorID = mockUser.Remote.ID
				mockAPI.On("KVSet", "sub_0c47c5b7e2a88ec9256c8ac0e71b0f6e", mock.Anything).Return(&model.AppError{Message: "Failed to store subscription"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(t, err, "Failed to store subscription")
			},
		},
		{
			name: "Error storing user settings",
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger) {
				mockAPI.ExpectedCalls = nil
				mockAPI.On("KVSet", "sub_0c47c5b7e2a88ec9256c8ac0e71b0f6e", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(&model.AppError{Message: "Failed to store user settings"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(t, err, "Failed to store user settings")
			},
		},
		{
			name: "Subscription successfully stored",
			setup: func(mockAPI *testutil.MockPluginAPI, mockLogger *mock_bot.MockLogger, mockLoggerWith *mock_bot.MockLogger) {
				mockAPI.ExpectedCalls = nil
				mockAPI.On("KVSet", "sub_0c47c5b7e2a88ec9256c8ac0e71b0f6e", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
				mockLogger.EXPECT().With(gomock.Any()).Return(mockLoggerWith).Times(1)
				mockLoggerWith.EXPECT().Debugf("store: stored mattermost user subscription.").Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI, store, mockLogger, mockLoggerWith, _ := GetMockSetup(t)
			tt.setup(mockAPI, mockLogger, mockLoggerWith)

			err := store.StoreUserSubscription(mockUser, mockSubscription)

			tt.assertions(t, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestDeleteUserSubscription(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI, *mock_bot.MockLogger, *mock_bot.MockLogger)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error deleting subscription",
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger) {
				mockAPI.On("KVDelete", "sub_0c47c5b7e2a88ec9256c8ac0e71b0f6e").Return(&model.AppError{Message: "Failed to delete subscription"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(t, err, "Failed to delete subscription")
			},
		},
		{
			name: "Error updating user settings",
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger) {
				mockAPI.ExpectedCalls = nil
				mockAPI.On("KVDelete", "sub_0c47c5b7e2a88ec9256c8ac0e71b0f6e").Return(nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(&model.AppError{Message: "Failed to update user settings"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(t, err, "Failed to update user settings")
			},
		},
		{
			name: "Subscription successfully deleted",
			setup: func(mockAPI *testutil.MockPluginAPI, mockLogger *mock_bot.MockLogger, mockLoggerWith *mock_bot.MockLogger) {
				mockAPI.ExpectedCalls = nil
				mockAPI.On("KVDelete", "sub_0c47c5b7e2a88ec9256c8ac0e71b0f6e").Return(nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "mmuid_e138a0f218087f9324d8c77f87d5f3a0", mock.Anything).Return(nil).Times(1)
				mockLogger.EXPECT().With(gomock.Any()).Return(mockLoggerWith).Times(1)
				mockLoggerWith.EXPECT().Debugf("store: deleted mattermost user subscription.").Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI, store, mockLogger, mockLoggerWith, _ := GetMockSetup(t)
			tt.setup(mockAPI, mockLogger, mockLoggerWith)

			mockUser := GetMockUser()
			err := store.DeleteUserSubscription(mockUser, MockSubscriptionID)

			tt.assertions(t, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

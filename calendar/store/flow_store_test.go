package store

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/testutil"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/tracker/mock_tracker"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestSetProperty(t *testing.T) {
	mockUser := GetMockUserWithSettings()
	mockUserJSON, err := json.Marshal(*mockUser)
	require.NoError(t, err)

	tests := []struct {
		name         string
		propertyName string
		value        interface{}
		setup        func(*testutil.MockPluginAPI, *mock_tracker.MockTracker)
		assertions   func(*testing.T, error)
	}{
		{
			name:         "Error loading user",
			propertyName: UpdateStatusPropertyName,
			value:        "online",
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(nil, &model.AppError{Message: "Error loading user"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "Error loading user")
			},
		},
		{
			name:         "Set UpdateStatusPropertyName successfully",
			propertyName: UpdateStatusPropertyName,
			value:        "online",
			setup: func(mockAPI *testutil.MockPluginAPI, mockTracker *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "mmuid_0404eb7ac36366cbc447d63a3acd7a5d", mock.Anything).Return(nil).Times(1)
				mockTracker.EXPECT().TrackAutomaticStatusUpdate("mockUserID", "online", "flow").Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:         "Set GetConfirmationPropertyName successfully",
			propertyName: GetConfirmationPropertyName,
			value:        true,
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "mmuid_0404eb7ac36366cbc447d63a3acd7a5d", mock.Anything).Return(nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:         "Set SetCustomStatusPropertyName successfully",
			propertyName: SetCustomStatusPropertyName,
			value:        false,
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "mmuid_0404eb7ac36366cbc447d63a3acd7a5d", mock.Anything).Return(nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:         "Set ReceiveUpcomingEventReminderName successfully",
			propertyName: ReceiveUpcomingEventReminderName,
			value:        false,
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "mmuid_0404eb7ac36366cbc447d63a3acd7a5d", mock.Anything).Return(nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:         "Invalid property name",
			propertyName: "mockPropertyName",
			value:        false,
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.EqualError(t, err, "property mockPropertyName not found")
			},
		},
		{
			name:         "Error storing user",
			propertyName: SetCustomStatusPropertyName,
			value:        true,
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(&model.AppError{Message: "Error storing user"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "Error storing user")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI, store, _, _, mockTracker := GetMockSetup(t)
			tt.setup(mockAPI, mockTracker)

			err := store.SetProperty(MockUserID, tt.propertyName, tt.value)

			tt.assertions(t, err)

			mockAPI.AssertExpectations(t)
		})
	}
}

func TestSetPostID(t *testing.T) {
	mockUser := GetMockUserWithSettings()
	mockUserJSON, err := json.Marshal(*mockUser)
	require.NoError(t, err)

	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error loading user",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(nil, &model.AppError{Message: "Error loading user"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "Error loading user")
			},
		},
		{
			name: "Set PostID successfully",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "mmuid_0404eb7ac36366cbc447d63a3acd7a5d", mock.Anything).Return(nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "Set PostID for user with nil PostIDs map",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockUser.WelcomeFlowStatus.PostIDs = nil
				mockUserJSON, _ = json.Marshal(mockUser)
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "mmuid_0404eb7ac36366cbc447d63a3acd7a5d", mock.Anything).Return(nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "Error storing user",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(&model.AppError{Message: "Error storing user"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "Error storing user")
			},
		},
	}
	for _, tt := range tests {
		mockAPI, store, _, _, _ := GetMockSetup(t)
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(mockAPI)

			err := store.SetPostID(MockUserID, "welcomePost", MockPostID)

			tt.assertions(t, err)

			mockAPI.AssertExpectations(t)
		})
	}
}

func TestGetPostID(t *testing.T) {
	mockUser := GetMockUserWithSettings()
	mockUserJSON, err := json.Marshal(*mockUser)
	require.NoError(t, err)

	tests := []struct {
		name         string
		propertyName string
		setup        func(*testutil.MockPluginAPI)
		assertions   func(*testing.T, string, error)
	}{
		{
			name:         "Error loading user",
			propertyName: "welcomePost",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(nil, &model.AppError{Message: "Error loading user"}).Times(1)
			},
			assertions: func(t *testing.T, postID string, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "Error loading user")
				require.Empty(t, postID)
			},
		},
		{
			name:         "PostID retrieved successfully",
			propertyName: "welcomePost",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
			},
			assertions: func(t *testing.T, postID string, err error) {
				require.NoError(t, err)
				require.Equal(t, MockPostID, postID)
			},
		},
		{
			name:         "PostID does not exist",
			propertyName: "nonExistentPost",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockUser.WelcomeFlowStatus.PostIDs = map[string]string{}
				mockUserJSON, _ = json.Marshal(mockUser)
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
			},
			assertions: func(t *testing.T, postID string, err error) {
				require.NoError(t, err)
				require.Empty(t, postID)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI, store, _, _, _ := GetMockSetup(t)
			tt.setup(mockAPI)

			postID, err := store.GetPostID(MockUserID, tt.propertyName)

			tt.assertions(t, postID, err)

			mockAPI.AssertExpectations(t)
		})
	}
}

func TestRemovePostID(t *testing.T) {
	mockUser := GetMockUserWithSettings()
	mockUserJSON, err := json.Marshal(*mockUser)
	require.NoError(t, err)

	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error loading user",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(nil, &model.AppError{Message: "Error loading user"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "Error loading user")
			},
		},
		{
			name: "Remove PostID successfully",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "mmuid_0404eb7ac36366cbc447d63a3acd7a5d", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "Error storing user",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(&model.AppError{Message: "Error storing user"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "Error storing user")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI, store, _, _, _ := GetMockSetup(t)
			tt.setup(mockAPI)

			err := store.RemovePostID(MockUserID, "welcomePost")

			tt.assertions(t, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestGetCurrentStep(t *testing.T) {
	mockUser := GetMockUserWithSettings()
	mockUserJSON, err := json.Marshal(*mockUser)
	require.NoError(t, err)

	tests := []struct {
		name       string
		userID     string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, int, error)
	}{
		{
			name:   "Error loading user",
			userID: "mockUserID",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(nil, &model.AppError{Message: "Error loading user"}).Times(1)
			},
			assertions: func(t *testing.T, step int, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "Error loading user")
				require.Equal(t, 0, step)
			},
		},
		{
			name:   "Get current step successfully",
			userID: "mockUserID",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
			},
			assertions: func(t *testing.T, step int, err error) {
				require.NoError(t, err)
				require.Equal(t, 3, step)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI, store, _, _, _ := GetMockSetup(t)
			tt.setup(mockAPI)

			step, err := store.GetCurrentStep(tt.userID)

			tt.assertions(t, step, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestSetCurrentStep(t *testing.T) {
	mockUser := GetMockUserWithSettings()
	mockUserJSON, err := json.Marshal(*mockUser)
	require.NoError(t, err)

	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error loading user",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(nil, &model.AppError{Message: "Error loading user"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "Error loading user")
			},
		},
		{
			name: "Error storing user",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(&model.AppError{Message: "Error storing user"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "Error storing user")
			},
		},
		{
			name: "Set current step successfully",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "mmuid_0404eb7ac36366cbc447d63a3acd7a5d", mock.Anything).Return(nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI, store, _, _, _ := GetMockSetup(t)
			tt.setup(mockAPI)

			err := store.SetCurrentStep(MockUserID, 2)

			tt.assertions(t, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestDeleteCurrentStep(t *testing.T) {
	mockUser := GetMockUserWithSettings()
	mockUserJSON, err := json.Marshal(*mockUser)
	require.NoError(t, err)

	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error loading user",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(nil, &model.AppError{Message: "Error loading user"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "Error loading user")
			},
		},
		{
			name: "Error storing user",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(&model.AppError{Message: "Error storing user"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "Error storing user")
			},
		},
		{
			name: "Delete current step successfully",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "mmuid_0404eb7ac36366cbc447d63a3acd7a5d", mock.Anything).Return(nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI, store, _, _, _ := GetMockSetup(t)
			tt.setup(mockAPI)
			err := store.DeleteCurrentStep(MockUserID)
			tt.assertions(t, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

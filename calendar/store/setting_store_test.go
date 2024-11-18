package store

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/testutil"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/tracker/mock_tracker"
)

func TestSetSetting(t *testing.T) {
	mockUser := GetMockUser()
	mockUserJSON, err := json.Marshal(mockUser)
	require.NoError(t, err)

	tests := []struct {
		name       string
		settingID  string
		value      interface{}
		setup      func(*testutil.MockPluginAPI, *mock_tracker.MockTracker)
		assertions func(*testing.T, error)
	}{
		{
			name:      "Error loading user",
			settingID: MockSettingID,
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(nil, &model.AppError{Message: "Error loading user"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.EqualError(t, err, "failed plugin KVGet: Error loading user")
			},
		},
		{
			name:      "error setting UpdateStatusFromOptionsSetting",
			settingID: UpdateStatusFromOptionsSettingID,
			value:     1,
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.EqualError(t, err, "cannot read value 1 for setting update_status_from_options (expecting string)")
			},
		},
		{
			name:      "Set UpdateStatusFromOptionsSetting",
			settingID: UpdateStatusFromOptionsSettingID,
			value:     "available",
			setup: func(mockAPI *testutil.MockPluginAPI, mockTracker *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "mmuid_e138a0f218087f9324d8c77f87d5f3a0", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
				mockTracker.EXPECT().TrackAutomaticStatusUpdate(MockUserID, "available", "settings").Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:      "error setting GetConfirmationSettingID",
			settingID: GetConfirmationSettingID,
			value:     1,
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.EqualError(t, err, "cannot read value 1 for setting get_confirmation (expecting bool)")
			},
		},
		{
			name:      "Set GetConfirmationSettingID",
			settingID: GetConfirmationSettingID,
			value:     true,
			setup: func(mockAPI *testutil.MockPluginAPI, mockTracker *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "mmuid_e138a0f218087f9324d8c77f87d5f3a0", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
				mockTracker.EXPECT().TrackAutomaticStatusUpdate(MockUserID, "available", "settings").Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:      "error setting SetCustomStatusSettingID",
			settingID: SetCustomStatusSettingID,
			value:     1,
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.EqualError(t, err, "cannot read value 1 for setting set_custom_status (expecting bool)")
			},
		},
		{
			name:      "Set SetCustomStatusSettingID",
			settingID: SetCustomStatusSettingID,
			value:     true,
			setup: func(mockAPI *testutil.MockPluginAPI, mockTracker *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "mmuid_e138a0f218087f9324d8c77f87d5f3a0", mock.Anything).Return(nil).Times(1)
				mockTracker.EXPECT().TrackAutomaticStatusUpdate(MockUserID, "available", "settings").Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:      "error setting ReceiveRemindersSettingID",
			settingID: ReceiveRemindersSettingID,
			value:     1,
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.EqualError(t, err, "cannot read value 1 for setting get_reminders (expecting bool)")
			},
		},
		{
			name:      "Set ReceiveRemindersSettingID",
			settingID: ReceiveRemindersSettingID,
			value:     true,
			setup: func(mockAPI *testutil.MockPluginAPI, mockTracker *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "mmuid_e138a0f218087f9324d8c77f87d5f3a0", mock.Anything).Return(nil).Times(1)
				mockTracker.EXPECT().TrackAutomaticStatusUpdate(MockUserID, "available", "settings").Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:      "Set DailySummarySettingID",
			settingID: DailySummarySettingID,
			value:     MockDailySummarySetting,
			setup: func(mockAPI *testutil.MockPluginAPI, mockTracker *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(nil).Times(1)
				mockAPI.On("KVSet", "mmuid_e138a0f218087f9324d8c77f87d5f3a0", mock.Anything).Return(nil).Times(1)
				mockTracker.EXPECT().TrackAutomaticStatusUpdate(MockUserID, "available", "settings").Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:      "invalid setting ID",
			settingID: "invalidSettingID",
			value:     MockDailySummarySetting,
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.EqualError(t, err, "setting invalidSettingID not found")
			},
		},
		{
			name:      "Error storing updated user",
			settingID: UpdateStatusFromOptionsSettingID,
			value:     "available",
			setup: func(mockAPI *testutil.MockPluginAPI, mockTracker *mock_tracker.MockTracker) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
				mockAPI.On("KVSet", "user_c3b5020d58a049787bc969768465b890", mock.Anything).Return(&model.AppError{Message: "Error storing user"}).Times(1)
				mockTracker.EXPECT().TrackAutomaticStatusUpdate(MockUserID, "available", "settings").Times(1)
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

			err := store.SetSetting(MockUserID, tt.settingID, tt.value)

			tt.assertions(t, err)

			mockAPI.AssertExpectations(t)
		})
	}
}

func TestGetSetting(t *testing.T) {
	mockUser := GetMockUserWithSettings()
	mockUserJSON, err := json.Marshal(*mockUser)
	require.NoError(t, err)

	tests := []struct {
		name       string
		settingID  string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, interface{}, error)
	}{
		{
			name:      "Error loading settings",
			settingID: MockSettingID,
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(nil, &model.AppError{Message: "Error loading settings"}).Times(1)
			},
			assertions: func(t *testing.T, setting interface{}, err error) {
				require.Error(t, err)
				require.Nil(t, setting)
				require.EqualError(t, err, "failed plugin KVGet: Error loading settings")
			},
		},
		{
			name:      "Get UpdateStatusFromOptionsSetting",
			settingID: UpdateStatusFromOptionsSettingID,
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
			},
			assertions: func(t *testing.T, setting interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "available", setting)
			},
		},
		{
			name:      "Get GetConfirmationSetting",
			settingID: GetConfirmationSettingID,
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
			},
			assertions: func(t *testing.T, setting interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, true, setting)
			},
		},
		{
			name:      "Get SetCustomStatusSetting",
			settingID: SetCustomStatusSettingID,
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
			},
			assertions: func(t *testing.T, setting interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, false, setting)
			},
		},
		{
			name:      "Get ReceiveRemindersSetting",
			settingID: ReceiveRemindersSettingID,
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
			},
			assertions: func(t *testing.T, setting interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, true, setting)
			},
		},
		{
			name:      "Get DailySummary",
			settingID: DailySummarySettingID,
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
			},
			assertions: func(t *testing.T, setting interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &DailySummaryUserSettings{PostTime: "10:00AM"}, setting)
			},
		},
		{
			name:      "invalid settingID",
			settingID: "invalidSettingID",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "user_ed8ba8dcdc37081824b09b84f8e061e6").Return(mockUserJSON, nil).Times(1)
			},
			assertions: func(t *testing.T, setting interface{}, err error) {
				require.EqualError(t, err, "setting invalidSettingID not found")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI, store, _, _, _ := GetMockSetup(t)
			tt.setup(mockAPI)

			setting, err := store.GetSetting(MockUserID, tt.settingID)

			tt.assertions(t, setting, err)

			mockAPI.AssertExpectations(t)
		})
	}
}

func TestDefaultDailySummaryUserSettings(t *testing.T) {
	dailySummaryUserSettings := DefaultDailySummaryUserSettings()

	require.Equal(t, "8:00AM", dailySummaryUserSettings.PostTime)
	require.Equal(t, "Eastern Standard Time", dailySummaryUserSettings.Timezone)
	require.Equal(t, false, dailySummaryUserSettings.Enable)
}

func TestSetPanelPostID(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error storing panel postID",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVSet", "settings_panel_ed8ba8dcdc37081824b09b84f8e061e6", mock.Anything).Return(&model.AppError{Message: "Failed to store panel postID"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(t, err, "Failed to store panel postID")
			},
		},
		{
			name: "Successful Stored panel postID",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVSet", "settings_panel_ed8ba8dcdc37081824b09b84f8e061e6", mock.Anything).Return(nil).Times(1)
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

			err := store.SetPanelPostID(MockUserID, MockPostID)

			tt.assertions(t, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestGetPanelPostID(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, string, error)
	}{
		{
			name: "Error loading panel postID",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "settings_panel_0c47c5b7e2a88ec9256c8ac0e71b0f6e").Return(nil, &model.AppError{Message: "Error loading panel postID"}).Times(1)
			},
			assertions: func(t *testing.T, panelPostID string, err error) {
				require.Error(t, err)
				require.Equal(t, "", panelPostID)
				require.EqualError(t, err, "failed plugin KVGet: Error loading panel postID")
			},
		},
		{
			name: "Success loading panel postID",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "settings_panel_0c47c5b7e2a88ec9256c8ac0e71b0f6e").Return([]byte(MockPostID), nil).Times(1)
			},
			assertions: func(t *testing.T, panelPostID string, err error) {
				require.NoError(t, err)
				require.Equal(t, MockPostID, panelPostID)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI, store, _, _, _ := GetMockSetup(t)
			tt.setup(mockAPI)

			panelPostID, err := store.GetPanelPostID(MockSubscriptionID)

			tt.assertions(t, panelPostID, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestDeletePanelPostID(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error deleting panel post id",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVDelete", "settings_panel_ed8ba8dcdc37081824b09b84f8e061e6").Return(&model.AppError{Message: "Failed to delete panel post id"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(t, err, "Failed to delete panel post id")
			},
		},
		{
			name: "Successful Delete panel post id",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVDelete", "settings_panel_ed8ba8dcdc37081824b09b84f8e061e6").Return(nil).Times(1)
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

			err := store.DeletePanelPostID(MockUserID)

			tt.assertions(t, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

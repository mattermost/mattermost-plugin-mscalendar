package store

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/testutil"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot/mock_bot"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestLoadUserEvent(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, *Event, error)
	}{
		{
			name: "Error loading event",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "ev_ff63e69da944334bfa44f98fe45e3c0c").Return(nil, &model.AppError{Message: "Event not found"}).Times(1)
			},
			assertions: func(t *testing.T, event *Event, err error) {
				require.Nil(t, event)
				require.EqualError(t, err, "failed plugin KVGet: Event not found")
			},
		},
		{
			name: "Successful Load",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "ev_ff63e69da944334bfa44f98fe45e3c0c").Return([]byte(`{"PluginVersion":"1.0","Remote":{"ID":"mockRemoteID"}}`), nil).Times(1)
			},
			assertions: func(t *testing.T, event *Event, err error) {
				require.NoError(t, err)
				require.Equal(t, "1.0", event.PluginVersion)
				require.Equal(t, MockRemoteID, event.Remote.ID)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI, store, _, _, _ := GetMockSetup(t)
			tt.setup(mockAPI)

			event, err := store.LoadUserEvent(MockUserID, MockEventID)

			tt.assertions(t, event, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestAddLinkedChannelToEvent(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error loading event metadata",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "ev_cf7c446273a2f147fa59573564da6b75").Return(nil, &model.AppError{Message: "Metadata not found"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "Metadata not found")
			},
		},
		{
			name: "Successful addition of linked channel",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "ev_cf7c446273a2f147fa59573564da6b75").Return(nil, nil).Times(1)
				mockAPI.On("KVSet", "ev_cf7c446273a2f147fa59573564da6b75", mock.Anything).Return(nil).Times(1)
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

			err := store.AddLinkedChannelToEvent(MockEventID, MockChannelID)

			tt.assertions(t, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestDeleteLinkedChannelFromEvent(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error loading event metadata",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "ev_cf7c446273a2f147fa59573564da6b75").Return(nil, &model.AppError{Message: "Metadata not found"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "Metadata not found")
			},
		},
		{
			name: "Channel ID not present",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "ev_cf7c446273a2f147fa59573564da6b75").Return([]byte(`{"LinkedChannelIDs":{"otherChannelID":{}}}`), nil).Times(1)
				mockAPI.On("KVSet", "ev_cf7c446273a2f147fa59573564da6b75", mock.Anything).Return(nil).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "Error storing updated metadata",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "ev_cf7c446273a2f147fa59573564da6b75").Return([]byte(`{"LinkedChannelIDs":{"mockChannelID":{}}}`), nil).Times(1)
				mockAPI.On("KVSet", "ev_cf7c446273a2f147fa59573564da6b75", mock.Anything).Return(&model.AppError{Message: "Failed to store metadata"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(t, err, "Failed to store metadata")
			},
		},
		{
			name: "Successful deletion of linked channel",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "ev_cf7c446273a2f147fa59573564da6b75").Return([]byte(`{"LinkedChannelIDs":{"mockChannelID":{}}}`), nil).Times(1)
				mockAPI.On("KVSet", "ev_cf7c446273a2f147fa59573564da6b75", mock.Anything).Return(nil).Times(1)
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

			err := store.DeleteLinkedChannelFromEvent(MockEventID, MockChannelID)

			tt.assertions(t, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestStoreEventMetadata(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error storing event metadata",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVSet", "ev_cf7c446273a2f147fa59573564da6b75", mock.Anything).Return(&model.AppError{Message: "Failed to store metadata"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "error storing event metadata")
			},
		},
		{
			name: "Successful store of event metadata",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVSet", "ev_cf7c446273a2f147fa59573564da6b75", mock.Anything).Return(nil).Times(1)
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

			eventMeta := &EventMetadata{
				LinkedChannelIDs: map[string]struct{}{
					MockChannelID: {},
				},
			}
			err := store.StoreEventMetadata(MockEventID, eventMeta)

			tt.assertions(t, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestLoadEventMetadata(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, *EventMetadata, error)
	}{
		{
			name: "Error loading event metadata",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "ev_cf7c446273a2f147fa59573564da6b75").Return(nil, &model.AppError{Message: "Failed to load event metadata"}).Times(1)
			},
			assertions: func(t *testing.T, eventMeta *EventMetadata, err error) {
				require.Nil(t, eventMeta)
				require.ErrorContains(t, err, "Failed to load event metadata")
			},
		},
		{
			name: "Successful load of event metadata",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "ev_cf7c446273a2f147fa59573564da6b75").Return([]byte(`{"LinkedChannelIDs":{"mockChannelID":{}}}`), nil).Times(1)
			},
			assertions: func(t *testing.T, eventMeta *EventMetadata, err error) {
				require.NoError(t, err)
				require.Contains(t, eventMeta.LinkedChannelIDs, "mockChannelID")
			},
		},
		{
			name: "Event metadata not found",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVGet", "ev_cf7c446273a2f147fa59573564da6b75").Return(nil, nil).Times(1)
			},
			assertions: func(t *testing.T, eventMeta *EventMetadata, err error) {
				require.ErrorContains(t, err, "not found")
				require.Nil(t, eventMeta)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI, store, _, _, _ := GetMockSetup(t)
			tt.setup(mockAPI)

			eventMeta, err := store.LoadEventMetadata(MockEventID)

			tt.assertions(t, eventMeta, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestDeleteEventMetadata(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error deleting event metadata",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVDelete", "ev_cf7c446273a2f147fa59573564da6b75").Return(&model.AppError{Message: "Failed to delete event metadata"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "Failed to delete event metadata")
			},
		},
		{
			name: "Successful deletion of event metadata",
			setup: func(mockAPI *testutil.MockPluginAPI) {
				mockAPI.On("KVDelete", "ev_cf7c446273a2f147fa59573564da6b75").Return(nil).Times(1)
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

			err := store.DeleteEventMetadata(MockEventID)

			tt.assertions(t, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestStoreUserEvent(t *testing.T) {
	mockEvent := GetMockEvent()

	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI, *mock_bot.MockLogger, *mock_bot.MockLogger)
		assertions func(*testing.T, error)
	}{
		{
			name: "Store expired event",
			setup: func(_ *testutil.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger) {
				mockEvent.Remote.End = &remote.DateTime{DateTime: "2006-01-02T15:04:05"}
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "Error storing user event",
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger) {
				mockEvent.Remote.End = remote.NewDateTime(time.Now(), "UTC")
				mockAPI.On("KVSetWithExpiry", "ev_ad2104c3b0ad765e6e9e03857a3348a5", mock.Anything, mock.AnythingOfType("int64")).Return(&model.AppError{Message: "Failed to store user event"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(t, err, "Failed to store user event")
			},
		},
		{
			name: "Successful store user event",
			setup: func(mockAPI *testutil.MockPluginAPI, mockLogger *mock_bot.MockLogger, mockLoggerWith *mock_bot.MockLogger) {
				mockEvent.Remote.End = remote.NewDateTime(time.Now(), "UTC")
				mockAPI.On("KVSetWithExpiry", "ev_ad2104c3b0ad765e6e9e03857a3348a5", mock.Anything, mock.AnythingOfType("int64")).Return(nil).Times(1)
				mockLogger.EXPECT().With(gomock.Any()).Return(mockLoggerWith).Times(1)
				mockLoggerWith.EXPECT().Debugf("store: stored user event.").Times(1)
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

			err := store.StoreUserEvent(MockUserID, mockEvent)

			tt.assertions(t, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestDeleteUserEvent(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*testutil.MockPluginAPI, *mock_bot.MockLogger, *mock_bot.MockLogger)
		assertions func(*testing.T, error)
	}{
		{
			name: "Error deleting user event",
			setup: func(mockAPI *testutil.MockPluginAPI, _ *mock_bot.MockLogger, _ *mock_bot.MockLogger) {
				mockAPI.On("KVDelete", mock.Anything).Return(&model.AppError{Message: "Failed to delete event"}).Times(1)
			},
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(t, err, "Failed to delete event")
			},
		},
		{
			name: "Successful delete",
			setup: func(mockAPI *testutil.MockPluginAPI, mockLogger *mock_bot.MockLogger, mockLoggerWith *mock_bot.MockLogger) {
				mockAPI.On("KVDelete", mock.Anything).Return(nil).Times(1)
				mockLogger.EXPECT().With(gomock.Any()).Return(mockLoggerWith).Times(1)
				mockLoggerWith.EXPECT().Debugf("store: deleted event.").Times(1)
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

			err := store.DeleteUserEvent("mockUserID", "mockEventID")

			tt.assertions(t, err)
			mockAPI.AssertExpectations(t)
		})
	}
}

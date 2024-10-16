package api

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine/mock_plugin_api"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote/mock_remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/store/mock_store"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot/mock_bot"
)

const (
	MMUserIDHeader = "Mattermost-User-Id"

	MockUserID       = "mockUserID"
	MockPostID       = "mockPostID"
	MockOption       = "mockOption"
	MockEventID      = "mockEventID"
	MockRemoteUserID = "mockRemoteUserID"
)

// revive:disable-next-line:unexported-return
func GetMockSetup(t *testing.T) (*api, *mock_store.MockStore, *mock_bot.MockPoster, *mock_remote.MockRemote, *mock_plugin_api.MockPluginAPI, *mock_bot.MockLogger, *mock_bot.MockLogger, *mock_remote.MockClient) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_store.NewMockStore(ctrl)
	mockPoster := mock_bot.NewMockPoster(ctrl)
	mockRemote := mock_remote.NewMockRemote(ctrl)
	mockPluginAPI := mock_plugin_api.NewMockPluginAPI(ctrl)
	mockLogger := mock_bot.NewMockLogger(ctrl)
	mockLoggerWith := mock_bot.NewMockLogger(ctrl)
	mockClient := mock_remote.NewMockClient(ctrl)

	env := engine.Env{
		Dependencies: &engine.Dependencies{
			Store:     mockStore,
			Poster:    mockPoster,
			Remote:    mockRemote,
			PluginAPI: mockPluginAPI,
			Logger:    mockLogger,
		},
	}

	api := &api{
		Env:                   env,
		NotificationProcessor: engine.NewNotificationProcessor(env),
	}

	return api, mockStore, mockPoster, mockRemote, mockPluginAPI, mockLogger, mockLoggerWith, mockClient
}

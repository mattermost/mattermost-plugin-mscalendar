package store

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/testutil"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/tracker/mock_tracker"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot/mock_bot"
)

func GetMockSetup(t *testing.T) (*testutil.MockPluginAPI, Store, *mock_bot.MockLogger, *mock_bot.MockLogger, *mock_tracker.MockTracker) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mock_bot.NewMockLogger(ctrl)
	mockLoggerWith := mock_bot.NewMockLogger(ctrl)
	mockTracker := mock_tracker.NewMockTracker(ctrl)
	mockAPI := &testutil.MockPluginAPI{}
	store := NewPluginStore(mockAPI, mockLogger, mockTracker, false, nil)

	return mockAPI, store, mockLogger, mockLoggerWith, mockTracker
}

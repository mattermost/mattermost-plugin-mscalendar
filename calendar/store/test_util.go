package store

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/testutil"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/tracker/mock_tracker"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot/mock_bot"
)

var MockString = mock.AnythingOfType("string")
var MockByteValue = mock.MatchedBy(func(arg interface{}) bool {
	_, ok := arg.([]byte)
	return ok
})

func GetMockSetup(t *testing.T) (*testutil.MockPluginAPI, Store, *mock_bot.MockLogger, *mock_bot.MockLogger, *mock_tracker.MockTracker) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mock_bot.NewMockLogger(ctrl)
	mockPoster := mock_bot.NewMockPoster(ctrl)
	mockLoggerWith := mock_bot.NewMockLogger(ctrl)
	mockTracker := mock_tracker.NewMockTracker(ctrl)
	mockAPI := &testutil.MockPluginAPI{}
	store := NewPluginStore(mockAPI, mockLogger, mockPoster, mockTracker, false, nil)

	return mockAPI, store, mockLogger, mockLoggerWith, mockTracker
}

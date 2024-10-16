package api

import (
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine/mock_plugin_api"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
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

type MockNotificationProcessor struct {
	queue []*remote.Notification
	err   error
}

func (m *MockNotificationProcessor) Enqueue(notifications ...*remote.Notification) error {
	if m.err != nil {
		return m.err
	}
	m.queue = append(m.queue, notifications...)
	return nil
}

func (m *MockNotificationProcessor) Configure(env engine.Env) {}
func (m *MockNotificationProcessor) Quit()                    {}

type mockClient struct{}

func (m *mockClient) GetMe() (*remote.User, error) {
	return &remote.User{ID: "mock-user-id"}, nil
}

func (m *mockClient) GetEvent(remoteUserID, eventID string) (*remote.Event, error) { return nil, nil }
func (m *mockClient) GetCalendars(remoteUserID string) ([]*remote.Calendar, error) { return nil, nil }
func (m *mockClient) GetDefaultCalendarView(remoteUserID string, start, end time.Time) ([]*remote.Event, error) {
	return nil, nil
}
func (m *mockClient) DoBatchViewCalendarRequests(params []*remote.ViewCalendarParams) ([]*remote.ViewCalendarResponse, error) {
	return nil, nil
}
func (m *mockClient) GetMailboxSettings(remoteUserID string) (*remote.MailboxSettings, error) {
	return nil, nil
}

func (m *mockClient) CreateEvent(remoteUserID string, calendarEvent *remote.Event) (*remote.Event, error) {
	return nil, nil
}
func (m *mockClient) AcceptEvent(remoteUserID, eventID string) error            { return nil }
func (m *mockClient) DeclineEvent(remoteUserID, eventID string) error           { return nil }
func (m *mockClient) TentativelyAcceptEvent(remoteUserID, eventID string) error { return nil }
func (m *mockClient) GetEventsBetweenDates(remoteUserID string, start, end time.Time) ([]*remote.Event, error) {
	return nil, nil
}

func (m *mockClient) CreateMySubscription(notificationURL, remoteUserID string) (*remote.Subscription, error) {
	return nil, nil
}
func (m *mockClient) DeleteSubscription(sub *remote.Subscription) error { return nil }
func (m *mockClient) GetNotificationData(notification *remote.Notification) (*remote.Notification, error) {
	return nil, nil
}
func (m *mockClient) ListSubscriptions() ([]*remote.Subscription, error) { return nil, nil }
func (m *mockClient) RenewSubscription(notificationURL, remoteUserID string, sub *remote.Subscription) (*remote.Subscription, error) {
	return nil, nil
}

func (m *mockClient) GetSuperuserToken() (string, error) { return "", nil }
func (m *mockClient) CallFormPost(method, path string, in url.Values, out interface{}) ([]byte, error) {
	return nil, nil
}
func (m *mockClient) CallJSON(method, path string, in, out interface{}) ([]byte, error) {
	return nil, nil
}

func (m *mockClient) CreateCalendar(remoteUserID string, calendar *remote.Calendar) (*remote.Calendar, error) {
	return nil, nil
}
func (m *mockClient) DeleteCalendar(remoteUserID, calendarID string) error { return nil }
func (m *mockClient) FindMeetingTimes(remoteUserID string, params *remote.FindMeetingTimesParameters) (*remote.MeetingTimeSuggestionResults, error) {
	return nil, nil
}

var _ remote.Client = (*mockClient)(nil)

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

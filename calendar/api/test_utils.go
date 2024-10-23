package api

import (
	"fmt"
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
	MockChannelID    = "mockChannelID"

	ValidRequestBodyJSON = `{
		"all_day": false,
		"attendees": ["user1", "user2"],
		"date": "2024-10-17",
		"start_time": "10:00AM",
		"end_time": "11:00AM",
		"description": "Team sync meeting",
		"subject": "Team Sync",
		"channel_id": "mockChannelID"
	}`
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

// revive:disable-next-line:unexported-return
func GetMockCreateEventPayload(allDay bool, attendees []string, date, startTime, endTime, description, subject, location, channelID string) createEventPayload {
	return createEventPayload{
		AllDay:      allDay,
		Attendees:   attendees,
		Date:        date,
		StartTime:   startTime,
		EndTime:     endTime,
		Description: description,
		Subject:     subject,
		Location:    location,
		ChannelID:   channelID,
	}
}

func GetCurrentTimeRequestBodyJSON(channelID string) string {
	currentTime := time.Now()
	date := currentTime.Format("2006-01-02")
	startTime := currentTime.Add(time.Hour).Format("15:04")
	endTime := currentTime.Add(2 * time.Hour).Format("15:04")

	return fmt.Sprintf(`{
					"all_day": false,
					"attendees": [],
					"date": "%s",
					"start_time": "%s",
					"end_time": "%s",
					"description": "Discuss the quarterly results.",
					"subject": "Meeting with team",
					"location": "Conference Room",
					"channel_id": "%s"
				}`, date, startTime, endTime, channelID)
}

func GetMockRemoteEvent() *remote.Event {
	currentTime := time.Now()
	return &remote.Event{
		Start: &remote.DateTime{
			DateTime: currentTime.Add(time.Hour).Format("2006-01-02T15:04:05Z"),
			TimeZone: "UTC",
		},
		End: &remote.DateTime{
			DateTime: currentTime.Add(2 * time.Hour).Format("2006-01-02T15:04:05Z"),
			TimeZone: "UTC",
		},
		Subject:  "Meeting with team",
		Location: &remote.Location{DisplayName: "Conference Room"},
		Conference: &remote.Conference{
			URL:         "https://example.com/conference",
			Application: "Zoom",
		},
		Attendees: []*remote.Attendee{
			{EmailAddress: &remote.EmailAddress{Name: "John Doe", Address: "john.doe@example.com"}},
			{EmailAddress: &remote.EmailAddress{Name: "Jane Smith", Address: "jane.smith@example.com"}},
		},
	}
}

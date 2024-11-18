package store

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/testutil"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/tracker/mock_tracker"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot/mock_bot"
)

const (
	MockEventSubscriptionID      = "mockEventSubscriptionID"
	MockSubscriptionID           = "mockSubscriptionID"
	MockCreatorID                = "mockCreatorID"
	MockMMUsername               = "mockMMUsername"
	MockMMDisplayName            = "mockMMDisplayName"
	MockMMUserID                 = "mockMMUserID"
	MockUserID                   = "mockUserID"
	MockSettingID                = "mockSettingID"
	MockPostID                   = "mockPostID"
	MockRemoteID                 = "mockRemoteID"
	MockRemoteUserID             = "mockRemoteUserID"
	MockRemoteMail               = "mock@remote.com"
	MockEventID                  = "mockEventID"
	MockChannelID                = "mockChannelID"
	MockUserIndexJSON            = `[{"mm_id": "mockMMUserID"}]`
	InvalidMockUserIndexJSON     = `[{"mm_id": "invalidMockMMUserID"}]`
	MockRemoteJSON               = `{"remote": {"id": "mockRemoteID"}}`
	MockUserJSON                 = `[{"MattermostUserID":"mockMMUserID","RemoteID":"mockRemoteID"}]`
	MockUserDetailsWithEventJSON = `{"mm_id":"mockUserID","active_events": []}`
	MockState                    = "mockState"
	MockDailySummarySetting      = "mockDailySummarySetting"
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

func GetMockUser() *User {
	return &User{
		MattermostUserID:      MockMMUserID,
		MattermostUsername:    MockMMUsername,
		MattermostDisplayName: MockMMDisplayName,
		Settings: Settings{
			EventSubscriptionID: MockEventSubscriptionID,
		},
		Remote: &remote.User{
			ID:   MockRemoteID,
			Mail: MockRemoteMail,
		},
	}
}

func GetMockUserWithSettings() *User {
	return &User{
		MattermostUserID: MockMMUserID,
		Remote:           &remote.User{ID: MockRemoteUserID},
		WelcomeFlowStatus: WelcomeFlowStatus{
			Step: 3,
			PostIDs: map[string]string{
				"welcomePost": "mockPostID",
			},
		},
		Settings: Settings{
			DailySummary: &DailySummaryUserSettings{
				PostTime: "10:00AM",
			},
			EventSubscriptionID:               MockEventSubscriptionID,
			UpdateStatusFromOptions:           "available",
			GetConfirmation:                   true,
			ReceiveReminders:                  true,
			SetCustomStatus:                   false,
			UpdateStatus:                      false,
			ReceiveNotificationsDuringMeeting: true,
		},
	}
}

func GetMockSubscription() *Subscription {
	return &Subscription{
		Remote: &remote.Subscription{
			ID:        MockSubscriptionID,
			CreatorID: MockCreatorID,
		},
	}
}

func GetRemoteUserJSON(noOfUsers int) string {
	type RemoteUser struct {
		MMUsername string `json:"mm_username"`
		RemoteID   string `json:"remote_id"`
		MMID       string `json:"mm_id"`
		Email      string `json:"email"`
	}

	var users []RemoteUser
	for i := 1; i <= noOfUsers; i++ {
		user := RemoteUser{
			MMUsername: fmt.Sprintf("user%d", i),
			RemoteID:   fmt.Sprintf("remote%d", i),
			MMID:       fmt.Sprintf("user%d", i),
			Email:      fmt.Sprintf("user%d@example.com", i),
		}
		users = append(users, user)
	}

	result, _ := json.Marshal(users)
	return string(result)
}

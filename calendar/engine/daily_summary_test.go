package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine/mock_plugin_api"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote/mock_remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/store/mock_store"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/telemetry"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/tracker"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot/mock_bot"
)

func TestGetDaySummaryForUser(t *testing.T) {
	t.Run("declined events are excluded", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mock_store.NewMockStore(ctrl)
		poster := mock_bot.NewMockPoster(ctrl)
		mockRemote := mock_remote.NewMockRemote(ctrl)
		mockClient := mock_remote.NewMockClient(ctrl)
		mockPluginAPI := mock_plugin_api.NewMockPluginAPI(ctrl)

		logger := mock_bot.NewMockLogger(ctrl)
		env := Env{
			Dependencies: &Dependencies{
				Store:     mockStore,
				Logger:    logger,
				Poster:    poster,
				Remote:    mockRemote,
				PluginAPI: mockPluginAPI,
				Tracker:   tracker.New(telemetry.NewTracker(nil, "", "", "", "", "", telemetry.TrackerConfig{}, nil)),
			},
		}

		loc, err := time.LoadLocation("EST")
		require.Nil(t, err)
		hour, minute := 9, 0 // Time is "9:00AM"
		moment := makeTime(hour, minute, loc)

		user := NewUser("user1_mm_id")

		mockStore.EXPECT().LoadUser(user.MattermostUserID).Return(&store.User{
			MattermostUserID: user.MattermostUserID,
			Remote:           &remote.User{ID: "user1_remote_id"},
			Settings: store.Settings{
				DailySummary: &store.DailySummaryUserSettings{
					Enable:       true,
					PostTime:     "9:00AM",
					Timezone:     "Pacific Standard Time",
					LastPostTime: "",
				},
			},
		}, nil).Times(2)

		mockPluginAPI.EXPECT().GetMattermostUser(user.MattermostUserID)

		mockRemote.EXPECT().MakeClient(context.Background(), nil).Return(mockClient)

		mockClient.EXPECT().GetMailboxSettings("user1_remote_id").Return(&remote.MailboxSettings{
			TimeZone: "Pacific Standard Time",
		}, nil)

		mockClient.EXPECT().GetDefaultCalendarView("user1_remote_id", gomock.Any(), gomock.Any()).Return([]*remote.Event{
			{
				Subject: "The subject",
				Start:   remote.NewDateTime(moment, "Pacific Standard Time"),
				End:     remote.NewDateTime(moment.Add(2*time.Hour), "Pacific Standard Time"),
			},
			{
				Subject: "The subject for declined event",
				Start:   remote.NewDateTime(moment, "Pacific Standard Time"),
				End:     remote.NewDateTime(moment.Add(2*time.Hour), "Pacific Standard Time"),
				ResponseStatus: &remote.EventResponseStatus{
					Response: remote.EventResponseStatusDeclined,
				},
			},
		}, nil)

		mscalendar := New(env, "user1_mm_id")
		result, err := mscalendar.GetDaySummaryForUser(moment, user)
		require.NoError(t, err)

		require.Equal(t, result, `Times are shown in Pacific Standard Time
Wednesday February 12, 2020

| Time | Subject |
| :-- | :-- |
| 9:00AM - 11:00AM | [The subject]() |`)
	})
}

func TestProcessAllDailySummary(t *testing.T) {
	for _, tc := range []struct {
		runAssertions func(deps *Dependencies, client remote.Client)
		name          string
		err           string
	}{
		{
			name: "Error fetching index",
			err:  "index store error",
			runAssertions: func(deps *Dependencies, client remote.Client) {
				s := deps.Store.(*mock_store.MockStore)
				s.EXPECT().LoadUserIndex().Return(nil, errors.New("index store error"))
			},
		},
		{
			name: "No users",
			err:  "",
			runAssertions: func(deps *Dependencies, client remote.Client) {
				s := deps.Store.(*mock_store.MockStore)
				s.EXPECT().LoadUserIndex().Return(store.UserIndex{}, nil)
			},
		},
		{
			name: "Error fetching events",
			err:  "error fetching events",
			runAssertions: func(deps *Dependencies, client remote.Client) {
				s := deps.Store.(*mock_store.MockStore)
				s.EXPECT().LoadUserIndex().Return(store.UserIndex{{
					MattermostUserID: "user1_mm_id",
					RemoteID:         "user1_remote_id",
				}}, nil)

				s.EXPECT().LoadUser("user1_mm_id").Return(&store.User{
					MattermostUserID: "user1_mm_id",
					Remote:           &remote.User{ID: "user1_remote_id"},
					Settings: store.Settings{
						DailySummary: &store.DailySummaryUserSettings{
							Enable:       true,
							PostTime:     "9:00AM",
							Timezone:     "Eastern Standard Time",
							LastPostTime: "",
						},
					},
				}, nil)

				mockClient := client.(*mock_remote.MockClient)
				mockRemote := deps.Remote.(*mock_remote.MockRemote)
				mockRemote.EXPECT().MakeSuperuserClient(context.Background()).Return(mockClient, nil).Times(1)

				mockClient.EXPECT().DoBatchViewCalendarRequests(gomock.Any()).Return([]*remote.ViewCalendarResponse{}, errors.New("error fetching events"))
			},
		},
		{
			name: "User receives their daily summary",
			err:  "",
			runAssertions: func(deps *Dependencies, client remote.Client) {
				s := deps.Store.(*mock_store.MockStore)
				s.EXPECT().LoadUserIndex().Return(store.UserIndex{{
					MattermostUserID: "user1_mm_id",
					RemoteID:         "user1_remote_id",
				}, {
					MattermostUserID: "user2_mm_id",
					RemoteID:         "user2_remote_id",
				}, {
					MattermostUserID: "user3_mm_id",
					RemoteID:         "user3_remote_id",
				}}, nil)

				s.EXPECT().LoadUser("user1_mm_id").Return(&store.User{
					MattermostUserID: "user1_mm_id",
					Remote:           &remote.User{ID: "user1_remote_id"},
					Settings: store.Settings{
						DailySummary: &store.DailySummaryUserSettings{
							Enable:       true,
							PostTime:     "9:00AM",
							Timezone:     "Eastern Standard Time",
							LastPostTime: "",
						},
					},
				}, nil)

				s.EXPECT().LoadUser("user2_mm_id").Return(&store.User{
					MattermostUserID: "user2_mm_id",
					Remote:           &remote.User{ID: "user2_remote_id"},
					Settings: store.Settings{
						DailySummary: &store.DailySummaryUserSettings{
							Enable:       true,
							PostTime:     "6:00AM",
							Timezone:     "Pacific Standard Time",
							LastPostTime: "",
						},
					},
				}, nil)

				s.EXPECT().LoadUser("user3_mm_id").Return(&store.User{
					MattermostUserID: "user3_mm_id",
					Remote:           &remote.User{ID: "user3_remote_id"},
					Settings: store.Settings{
						DailySummary: &store.DailySummaryUserSettings{
							Enable:       true,
							PostTime:     "10:00AM", // should not receive summary
							Timezone:     "Pacific Standard Time",
							LastPostTime: "",
						},
					},
				}, nil)

				mockClient := client.(*mock_remote.MockClient)
				loc, err := time.LoadLocation("MST")
				require.Nil(t, err)
				hour, minute := 10, 0 // Time is "10:00AM"
				moment := makeTime(hour, minute, loc)
				mockClient.EXPECT().DoBatchViewCalendarRequests(gomock.Any()).Return([]*remote.ViewCalendarResponse{
					{RemoteUserID: "user1_remote_id", Events: []*remote.Event{}},
					{RemoteUserID: "user2_remote_id", Events: []*remote.Event{
						{
							Subject: "The subject",
							Start:   remote.NewDateTime(moment, "Mountain Standard Time"),
							End:     remote.NewDateTime(moment.Add(2*time.Hour), "Mountain Standard Time"),
							Location: &remote.Location{
								DisplayName: "https://zoom.us/j/123",
							},
						},
					}},
				}, nil)
				mockRemote := deps.Remote.(*mock_remote.MockRemote)
				mockRemote.EXPECT().MakeSuperuserClient(context.Background()).Return(mockClient, nil).Times(1)

				mockPoster := deps.Poster.(*mock_bot.MockPoster)
				gomock.InOrder(
					mockPoster.EXPECT().DM("user1_mm_id", "You have no upcoming events.").Return("postID1", nil).Times(1),
					mockPoster.EXPECT().DM("user2_mm_id", `Times are shown in Pacific Standard Time
Wednesday February 12, 2020

| Time | Subject |
| :-- | :-- |
| 9:00AM - 11:00AM | [The subject]() |`).Return("postID2", nil).Times(1),
				)

				s.EXPECT().StoreUser(gomock.Any()).Times(2).DoAndReturn(func(u *store.User) error {
					require.NotEmpty(t, u.Settings.DailySummary.LastPostTime)
					return nil
				})

				mockLogger := deps.Logger.(*mock_bot.MockLogger)
				mockLogger.EXPECT().Infof("Processed daily summary for %d users", 2)
			},
		},
		{
			name: "User receives their daily summary (individual data call)",
			err:  "",
			runAssertions: func(deps *Dependencies, client remote.Client) {
				s := deps.Store.(*mock_store.MockStore)
				mockRemote := deps.Remote.(*mock_remote.MockRemote)
				mockClient := client.(*mock_remote.MockClient)
				papi := deps.PluginAPI.(*mock_plugin_api.MockPluginAPI)

				loc, err := time.LoadLocation("MST")
				require.Nil(t, err)
				hour, minute := 10, 0 // Time is "10:00AM"
				moment := makeTime(hour, minute, loc)

				s.EXPECT().LoadUserIndex().Return(store.UserIndex{{
					MattermostUserID: "user1_mm_id",
					RemoteID:         "user1_remote_id",
				}, {
					MattermostUserID: "user2_mm_id",
					RemoteID:         "user2_remote_id",
				}, {
					MattermostUserID: "user3_mm_id",
					RemoteID:         "user3_remote_id",
				}}, nil)

				mockRemote.EXPECT().MakeSuperuserClient(context.Background()).Return(nil, remote.ErrSuperUserClientNotSupported).Times(1)

				s.EXPECT().LoadUser("user1_mm_id").Return(&store.User{
					MattermostUserID: "user1_mm_id",
					Remote:           &remote.User{ID: "user1_remote_id"},
					Settings: store.Settings{
						DailySummary: &store.DailySummaryUserSettings{
							Enable:       true,
							PostTime:     "9:00AM",
							Timezone:     "Eastern Standard Time",
							LastPostTime: "",
						},
					},
				}, nil).Times(3)

				s.EXPECT().LoadUser("user2_mm_id").Return(&store.User{
					MattermostUserID: "user2_mm_id",
					Remote:           &remote.User{ID: "user2_remote_id"},
					Settings: store.Settings{
						DailySummary: &store.DailySummaryUserSettings{
							Enable:       true,
							PostTime:     "6:00AM",
							Timezone:     "Pacific Standard Time",
							LastPostTime: "",
						},
					},
				}, nil).Times(3)

				s.EXPECT().LoadUser("user3_mm_id").Return(&store.User{
					MattermostUserID: "user3_mm_id",
					Remote:           &remote.User{ID: "user3_remote_id"},
					Settings: store.Settings{
						DailySummary: &store.DailySummaryUserSettings{
							Enable:       true,
							PostTime:     "10:00AM", // should not receive summary
							Timezone:     "Pacific Standard Time",
							LastPostTime: "",
						},
					},
				}, nil)

				papi.EXPECT().GetMattermostUser("user1_mm_id").Times(2)
				papi.EXPECT().GetMattermostUser("user2_mm_id").Times(2)

				mockClient.EXPECT().GetMailboxSettings("user1_remote_id").Return(&remote.MailboxSettings{
					TimeZone: "Eastern Standard Time",
				}, nil)
				mockClient.EXPECT().GetMailboxSettings("user2_remote_id").Return(&remote.MailboxSettings{
					TimeZone: "Pacific Standard Time",
				}, nil)

				mockRemote.EXPECT().MakeClient(context.Background(), gomock.Any()).Return(mockClient).Times(2)

				mockClient.EXPECT().GetDefaultCalendarView("user1_remote_id", gomock.Any(), gomock.Any()).Return([]*remote.Event{}, nil)
				mockClient.EXPECT().GetDefaultCalendarView("user2_remote_id", gomock.Any(), gomock.Any()).Return([]*remote.Event{
					{
						Subject: "The subject",
						Start:   remote.NewDateTime(moment, "Mountain Standard Time"),
						End:     remote.NewDateTime(moment.Add(2*time.Hour), "Mountain Standard Time"),
					},
				}, nil)

				mockPoster := deps.Poster.(*mock_bot.MockPoster)
				gomock.InOrder(
					mockPoster.EXPECT().DM("user1_mm_id", "You have no upcoming events.").Return("postID1", nil).Times(1),
					mockPoster.EXPECT().DM("user2_mm_id", `Times are shown in Pacific Standard Time
Wednesday February 12, 2020

| Time | Subject |
| :-- | :-- |
| 9:00AM - 11:00AM | [The subject]() |`).Return("postID2", nil).Times(1),
				)

				s.EXPECT().StoreUser(gomock.Any()).Times(2).DoAndReturn(func(u *store.User) error {
					require.NotEmpty(t, u.Settings.DailySummary.LastPostTime)
					return nil
				})

				mockLogger := deps.Logger.(*mock_bot.MockLogger)
				mockLogger.EXPECT().Infof("Processed daily summary for %d users", 2)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := mock_store.NewMockStore(ctrl)
			poster := mock_bot.NewMockPoster(ctrl)
			mockRemote := mock_remote.NewMockRemote(ctrl)
			mockClient := mock_remote.NewMockClient(ctrl)
			mockPluginAPI := mock_plugin_api.NewMockPluginAPI(ctrl)

			logger := mock_bot.NewMockLogger(ctrl)
			env := Env{
				Dependencies: &Dependencies{
					Store:     s,
					Logger:    logger,
					Poster:    poster,
					Remote:    mockRemote,
					PluginAPI: mockPluginAPI,
					Tracker:   tracker.New(telemetry.NewTracker(nil, "", "", "", "", "", telemetry.TrackerConfig{}, nil)),
				},
			}

			loc, err := time.LoadLocation("EST")
			require.Nil(t, err)
			hour, minute := 9, 0 // Time is "9:00AM"
			moment := makeTime(hour, minute, loc)

			tc.runAssertions(env.Dependencies, mockClient)
			mscalendar := New(env, "")
			err = mscalendar.ProcessAllDailySummary(moment)

			if tc.err != "" {
				require.Equal(t, tc.err, err.Error())
			} else {
				require.Nil(t, err)
			}

			require.NotNil(t, tc)
		})
	}
}

func TestShouldPostDailySummary(t *testing.T) {
	tests := []struct {
		name        string
		postTime    string
		timeZone    string
		enabled     bool
		shouldRun   bool
		shouldError bool
	}{
		{
			name:        "Disabled",
			enabled:     false,
			postTime:    "9:00AM",
			timeZone:    "Eastern Standard Time",
			shouldRun:   false,
			shouldError: false,
		},
		{
			name:        "Same timezone, wrong time",
			enabled:     true,
			postTime:    "8:00AM",
			timeZone:    "Eastern Standard Time",
			shouldRun:   false,
			shouldError: false,
		},
		{
			name:        "Same timezone, right time",
			enabled:     true,
			postTime:    "9:00AM",
			timeZone:    "Eastern Standard Time",
			shouldRun:   true,
			shouldError: false,
		},
		{
			name:        "Different timezone, wrong time",
			enabled:     true,
			postTime:    "9:00AM",
			timeZone:    "Mountain Standard Time",
			shouldRun:   false,
			shouldError: false,
		},
		{
			name:        "Different timezone, right time",
			enabled:     true,
			postTime:    "7:00AM",
			timeZone:    "Mountain Standard Time",
			shouldRun:   true,
			shouldError: false,
		},
		{
			name:        "Nepal timezone, wrong time",
			enabled:     true,
			postTime:    "7:00AM",
			timeZone:    "Nepal Standard Time",
			shouldRun:   false,
			shouldError: false,
		},
		{
			name:        "Nepal timezone, right time",
			enabled:     true,
			postTime:    "7:45PM",
			timeZone:    "Nepal Standard Time",
			shouldRun:   true,
			shouldError: false,
		},
		{
			enabled:     true,
			postTime:    "7:20FM", // Invalid time
			timeZone:    "Mountain Standard Time",
			shouldRun:   false,
			shouldError: true,
		},
		{
			enabled:     true,
			postTime:    "7:00AM",
			timeZone:    "Moon Time",
			shouldRun:   false,
			shouldError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			loc, err := time.LoadLocation("EST")
			require.Nil(t, err)

			dsum := &store.DailySummaryUserSettings{
				Enable:   tc.enabled,
				PostTime: tc.postTime,
				Timezone: tc.timeZone,
			}

			hour, minute := 9, 0 // Time is "9:00AM"
			moment := makeTime(hour, minute, loc)

			shouldRun, err := shouldPostDailySummary(dsum, moment)
			require.Equal(t, tc.shouldRun, shouldRun)
			if tc.shouldError {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func makeTime(hour, minute int, loc *time.Location) time.Time {
	return time.Date(2020, 2, 12, hour, minute, 0, 0, loc)
}

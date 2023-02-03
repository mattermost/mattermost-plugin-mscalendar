// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/mock_plugin_api"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote/mock_remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store/mock_store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot/mock_bot"
)

func TestSyncStatusAll(t *testing.T) {
	moment := time.Now().UTC()
	eventHash := "event_id " + moment.Format(time.RFC3339)
	busyEvent := &remote.Event{ICalUID: "event_id", Start: remote.NewDateTime(moment, "UTC"), ShowAs: "busy"}

	for name, tc := range map[string]struct {
		apiError            *remote.APIError
		currentStatus       string
		newStatus           string
		remoteEvents        []*remote.Event
		activeEvents        []string
		eventsToStore       []string
		currentStatusManual bool
		shouldLogError      bool
		getConfirmation     bool
	}{
		"Most common case, no events local or remote. No status change.": {
			remoteEvents:        []*remote.Event{},
			activeEvents:        []string{},
			currentStatus:       "online",
			currentStatusManual: true,
			newStatus:           "",
			eventsToStore:       nil,
			shouldLogError:      false,
			getConfirmation:     false,
		},
		"New remote event. Change status to DND.": {
			remoteEvents:        []*remote.Event{busyEvent},
			activeEvents:        []string{},
			currentStatus:       "online",
			currentStatusManual: true,
			newStatus:           "dnd",
			eventsToStore:       []string{eventHash},
			shouldLogError:      false,
			getConfirmation:     false,
		},
		"Locally stored event is finished. Change status to online.": {
			remoteEvents:        []*remote.Event{},
			activeEvents:        []string{eventHash},
			currentStatus:       "dnd",
			currentStatusManual: true,
			newStatus:           "online",
			eventsToStore:       []string{},
			shouldLogError:      false,
			getConfirmation:     false,
		},
		"Locally stored event is still happening. No status change.": {
			remoteEvents:        []*remote.Event{busyEvent},
			activeEvents:        []string{eventHash},
			currentStatus:       "dnd",
			currentStatusManual: true,
			newStatus:           "",
			eventsToStore:       nil,
			shouldLogError:      false,
			getConfirmation:     false,
		},
		"User has manually set themselves to online during event. Locally stored event is still happening, but we will ignore it. No status change.": {
			remoteEvents:        []*remote.Event{busyEvent},
			activeEvents:        []string{eventHash},
			currentStatus:       "online",
			currentStatusManual: true,
			newStatus:           "",
			eventsToStore:       nil,
			shouldLogError:      false,
			getConfirmation:     false,
		},
		"Ignore non-busy event": {
			remoteEvents:        []*remote.Event{{ID: "event_id_2", Start: remote.NewDateTime(moment, "UTC"), ShowAs: "free"}},
			activeEvents:        []string{},
			currentStatus:       "online",
			currentStatusManual: true,
			newStatus:           "",
			eventsToStore:       nil,
			shouldLogError:      false,
			getConfirmation:     false,
		},
		"Remote API error. Error should be logged": {
			remoteEvents:        nil,
			activeEvents:        []string{eventHash},
			currentStatus:       "online",
			currentStatusManual: true,
			newStatus:           "",
			eventsToStore:       nil,
			apiError:            &remote.APIError{Code: "403", Message: "Forbidden"},
			shouldLogError:      true,
			getConfirmation:     false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			env, client := makeStatusSyncTestEnv(ctrl)
			deps := env.Dependencies

			c, papi, s, logger := client.(*mock_remote.MockClient), deps.PluginAPI.(*mock_plugin_api.MockPluginAPI), deps.Store.(*mock_store.MockStore), deps.Logger.(*mock_bot.MockLogger)

			mockUser := &store.User{
				MattermostUserID: "user_mm_id",
				Remote: &remote.User{
					ID:   "user_remote_id",
					Mail: "user_email@example.com",
				},
				Settings:     store.Settings{UpdateStatus: true, GetConfirmation: tc.getConfirmation},
				ActiveEvents: tc.activeEvents,
			}
			s.EXPECT().LoadUser("user_mm_id").Return(mockUser, nil).Times(1)

			c.EXPECT().DoBatchViewCalendarRequests(gomock.Any()).Return([]*remote.ViewCalendarResponse{
				{Events: tc.remoteEvents, RemoteUserID: "user_remote_id", Error: tc.apiError},
			}, nil)

			papi.EXPECT().GetMattermostUserStatusesByIds([]string{"user_mm_id"}).Return([]*model.Status{{Status: tc.currentStatus, Manual: tc.currentStatusManual, UserId: "user_mm_id"}}, nil)

			if tc.newStatus == "" {
				papi.EXPECT().UpdateMattermostUserStatus("user_mm_id", gomock.Any()).Times(0)
			} else {
				if tc.currentStatusManual && !tc.getConfirmation ||
					tc.currentStatusManual && tc.currentStatus == "dnd" {
					if tc.newStatus == "dnd" {
						mockUser.LastStatus = tc.currentStatus
					}
					s.EXPECT().StoreUser(mockUser).Return(nil).Times(1)
				}
				papi.EXPECT().UpdateMattermostUserStatus("user_mm_id", tc.newStatus).Return(nil, nil)
			}

			if tc.eventsToStore == nil {
				s.EXPECT().StoreUserActiveEvents("user_mm_id", gomock.Any()).Return(nil).Times(0)
			} else {
				s.EXPECT().StoreUserActiveEvents("user_mm_id", tc.eventsToStore).Return(nil).Times(1)
			}

			if tc.shouldLogError {
				logger.EXPECT().Warnf("Error getting availability for %s. err=%s", "user_email@example.com", tc.apiError.Message).Times(1)
			} else {
				logger.EXPECT().Warnf(gomock.Any()).Times(0)
			}

			m := New(env, "")
			res, err := m.SyncAll()
			require.Nil(t, err)
			require.NotEmpty(t, res)
		})
	}
}

func TestSyncStatusUserConfig(t *testing.T) {
	for name, tc := range map[string]struct {
		runAssertions func(deps *Dependencies, client remote.Client)
		settings      store.Settings
	}{
		"UpdateStatus disabled": {
			settings: store.Settings{
				UpdateStatus: false,
			},
			runAssertions: func(deps *Dependencies, client remote.Client) {
				c := client.(*mock_remote.MockClient)
				c.EXPECT().DoBatchViewCalendarRequests(gomock.Any()).Times(0)
			},
		},
		"UpdateStatus enabled and GetConfirmation enabled": {
			settings: store.Settings{
				UpdateStatus:    true,
				GetConfirmation: true,
			},
			runAssertions: func(deps *Dependencies, client remote.Client) {
				c, papi, poster, s := client.(*mock_remote.MockClient), deps.PluginAPI.(*mock_plugin_api.MockPluginAPI), deps.Poster.(*mock_bot.MockPoster), deps.Store.(*mock_store.MockStore)
				moment := time.Now().UTC()
				busyEvent := &remote.Event{ICalUID: "event_id", Start: remote.NewDateTime(moment, "UTC"), ShowAs: "busy"}

				c.EXPECT().DoBatchViewCalendarRequests(gomock.Any()).Times(1).Return([]*remote.ViewCalendarResponse{
					{Events: []*remote.Event{busyEvent}, RemoteUserID: "user_remote_id"},
				}, nil)
				papi.EXPECT().GetMattermostUserStatusesByIds([]string{"user_mm_id"}).Return([]*model.Status{{Status: "online", Manual: true, UserId: "user_mm_id"}}, nil)

				s.EXPECT().StoreUser(gomock.Any()).Return(nil).Times(1)
				s.EXPECT().StoreUserActiveEvents("user_mm_id", []string{"event_id " + moment.Format(time.RFC3339)})
				poster.EXPECT().DMWithAttachments("user_mm_id", gomock.Any()).Times(1)
				papi.EXPECT().UpdateMattermostUserStatus("user_mm_id", gomock.Any()).Times(0)
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			env, client := makeStatusSyncTestEnv(ctrl)

			s := env.Dependencies.Store.(*mock_store.MockStore)
			s.EXPECT().LoadUser("user_mm_id").Return(&store.User{
				MattermostUserID: "user_mm_id",
				Remote: &remote.User{
					ID:   "user_remote_id",
					Mail: "user_email@example.com",
				},
				Settings: tc.settings,
			}, nil).Times(1)

			tc.runAssertions(env.Dependencies, client)

			mscalendar := New(env, "")
			_, err := mscalendar.SyncAll()
			require.Nil(t, err)
		})
	}
}

func TestReminders(t *testing.T) {
	for name, tc := range map[string]struct {
		apiError       *remote.APIError
		remoteEvents   []*remote.Event
		numReminders   int
		shouldLogError bool
	}{
		"Most common case, no remote events. No reminder.": {
			remoteEvents:   []*remote.Event{},
			numReminders:   0,
			shouldLogError: false,
		},
		"One remote event, but it is too far in the future.": {
			remoteEvents: []*remote.Event{
				{ICalUID: "event_id", Start: remote.NewDateTime(time.Now().Add(20*time.Minute).UTC(), "UTC"), End: remote.NewDateTime(time.Now().Add(45*time.Minute).UTC(), "UTC")},
			},
			numReminders:   0,
			shouldLogError: false,
		},
		"One remote event, but it is in the past.": {
			remoteEvents: []*remote.Event{
				{ICalUID: "event_id", Start: remote.NewDateTime(time.Now().Add(-15*time.Minute).UTC(), "UTC"), End: remote.NewDateTime(time.Now().Add(45*time.Minute).UTC(), "UTC")},
			},
			numReminders:   0,
			shouldLogError: false,
		},
		"One remote event, but it is to soon in the future. Reminder has already occurred.": {
			remoteEvents: []*remote.Event{
				{ICalUID: "event_id", Start: remote.NewDateTime(time.Now().Add(2*time.Minute).UTC(), "UTC"), End: remote.NewDateTime(time.Now().Add(45*time.Minute).UTC(), "UTC")},
			},
			numReminders:   0,
			shouldLogError: false,
		},
		"One remote event, and is in the range for the reminder. Reminder should occur.": {
			remoteEvents: []*remote.Event{
				{ICalUID: "event_id", Start: remote.NewDateTime(time.Now().Add(7*time.Minute).UTC(), "UTC"), End: remote.NewDateTime(time.Now().Add(45*time.Minute).UTC(), "UTC")},
			},
			numReminders:   1,
			shouldLogError: false,
		},
		"Two remote event, and are in the range for the reminder. Two reminders should occur.": {
			remoteEvents: []*remote.Event{
				{ICalUID: "event_id", Start: remote.NewDateTime(time.Now().Add(7*time.Minute).UTC(), "UTC"), End: remote.NewDateTime(time.Now().Add(45*time.Minute).UTC(), "UTC")},
				{ICalUID: "event_id", Start: remote.NewDateTime(time.Now().Add(7*time.Minute).UTC(), "UTC"), End: remote.NewDateTime(time.Now().Add(45*time.Minute).UTC(), "UTC")},
			},
			numReminders:   2,
			shouldLogError: false,
		},
		"Remote API Error. Error should be logged.": {
			remoteEvents:   []*remote.Event{},
			numReminders:   0,
			apiError:       &remote.APIError{Code: "403", Message: "Forbidden"},
			shouldLogError: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			env, client := makeStatusSyncTestEnv(ctrl)
			deps := env.Dependencies

			c, s, poster, logger := client.(*mock_remote.MockClient), deps.Store.(*mock_store.MockStore), deps.Poster.(*mock_bot.MockPoster), deps.Logger.(*mock_bot.MockLogger)

			loadUser := s.EXPECT().LoadUser("user_mm_id").Return(&store.User{
				MattermostUserID: "user_mm_id",
				Remote: &remote.User{
					ID:   "user_remote_id",
					Mail: "user_email@example.com",
				},
				Settings: store.Settings{ReceiveReminders: true},
			}, nil)
			c.EXPECT().DoBatchViewCalendarRequests(gomock.Any()).Return([]*remote.ViewCalendarResponse{
				{Events: tc.remoteEvents, RemoteUserID: "user_remote_id", Error: tc.apiError},
			}, nil)

			if tc.numReminders > 0 {
				poster.EXPECT().DM("user_mm_id", gomock.Any()).Times(tc.numReminders)
				loadUser.Times(2)
				c.EXPECT().GetMailboxSettings("user_remote_id").Times(1).Return(&remote.MailboxSettings{TimeZone: "UTC"}, nil)
			} else {
				poster.EXPECT().DM(gomock.Any(), gomock.Any()).Times(0)
				loadUser.Times(1)
			}

			if tc.shouldLogError {
				logger.EXPECT().Warnf("Error getting availability for %s. err=%s", "user_email@example.com", tc.apiError.Message).Times(1)
			} else {
				logger.EXPECT().Warnf(gomock.Any()).Times(0)
			}

			m := New(env, "")
			res, err := m.SyncAll()
			require.Nil(t, err)
			require.NotEmpty(t, res)
		})
	}
}

func makeStatusSyncTestEnv(ctrl *gomock.Controller) (Env, remote.Client) {
	s := mock_store.NewMockStore(ctrl)
	poster := mock_bot.NewMockPoster(ctrl)
	mockRemote := mock_remote.NewMockRemote(ctrl)
	mockClient := mock_remote.NewMockClient(ctrl)
	mockPluginAPI := mock_plugin_api.NewMockPluginAPI(ctrl)
	logger := mock_bot.NewMockLogger(ctrl)

	env := Env{
		Config: &config.Config{},
		Dependencies: &Dependencies{
			Store:     s,
			Logger:    logger,
			Poster:    poster,
			Remote:    mockRemote,
			PluginAPI: mockPluginAPI,
		},
	}

	s.EXPECT().LoadUserIndex().Return(store.UserIndex{
		&store.UserShort{
			MattermostUserID: "user_mm_id",
			RemoteID:         "user_remote_id",
			Email:            "user_email@example.com",
		},
	}, nil).Times(1)

	mockRemote.EXPECT().MakeSuperuserClient(context.Background()).Return(mockClient, nil)

	return env, mockClient
}

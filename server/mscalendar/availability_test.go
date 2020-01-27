// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/mock_mscalendar"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote/mock_remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store/mock_store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot/mock_bot"
	"github.com/mattermost/mattermost-server/v5/model"
)

func TestSyncStatusForAllUsers(t *testing.T) {
	for name, tc := range map[string]struct {
		sched         *remote.ScheduleInformation
		currentStatus string
		newStatus     string
	}{
		"User is free but dnd, mark user as online": {
			sched: &remote.ScheduleInformation{
				ScheduleID:       "some_email@example.com",
				AvailabilityView: "0",
			},
			currentStatus: "dnd",
			newStatus:     "online",
		},
		"User is busy but online, mark as dnd": {
			sched: &remote.ScheduleInformation{
				ScheduleID:       "some_email@example.com",
				AvailabilityView: "2",
			},
			currentStatus: "online",
			newStatus:     "dnd",
		},
		"User is free and online, do not change status": {
			sched: &remote.ScheduleInformation{
				ScheduleID:       "some_email@example.com",
				AvailabilityView: "0",
			},
			currentStatus: "online",
			newStatus:     "",
		},
		"User is busy and dnd, do not change status": {
			sched: &remote.ScheduleInformation{
				ScheduleID:       "some_email@example.com",
				AvailabilityView: "2",
			},
			currentStatus: "dnd",
			newStatus:     "",
		},
	} {
		t.Run(name, func(t *testing.T) {
			userStoreCtrl := gomock.NewController(t)
			defer userStoreCtrl.Finish()
			userStore := mock_store.NewMockUserStore(userStoreCtrl)

			oauthStoreCtrl := gomock.NewController(t)
			defer oauthStoreCtrl.Finish()
			oauthStore := mock_store.NewMockOAuth2StateStore(oauthStoreCtrl)

			subsStoreCtrl := gomock.NewController(t)
			defer subsStoreCtrl.Finish()
			subsStore := mock_store.NewMockSubscriptionStore(subsStoreCtrl)

			eventStoreCtrl := gomock.NewController(t)
			defer eventStoreCtrl.Finish()
			eventStore := mock_store.NewMockEventStore(eventStoreCtrl)

			conf := &config.Config{}

			posterCtrl := gomock.NewController(t)
			defer posterCtrl.Finish()
			poster := mock_bot.NewMockPoster(posterCtrl)

			logger := &bot.TestLogger{TB: t}

			remoteCtrl := gomock.NewController(t)
			defer remoteCtrl.Finish()
			mockRemote := mock_remote.NewMockRemote(remoteCtrl)

			clientCtrl := gomock.NewController(t)
			defer clientCtrl.Finish()
			mockClient := mock_remote.NewMockClient(clientCtrl)

			pluginAPICtrl := gomock.NewController(t)
			defer pluginAPICtrl.Finish()
			mockPluginAPI := mock_mscalendar.NewMockPluginAPI(pluginAPICtrl)

			apiConfig := Config{
				Config: conf,
				Dependencies: &Dependencies{
					UserStore:         userStore,
					OAuth2StateStore:  oauthStore,
					SubscriptionStore: subsStore,
					EventStore:        eventStore,
					Logger:            logger,
					Poster:            poster,
					Remote:            mockRemote,
					PluginAPI:         mockPluginAPI,
				},
			}

			userStore.EXPECT().LoadUserIndex().Return(store.UserIndex{
				&store.UserShort{
					MattermostUserID: "some_mm_id",
					RemoteID:         "some_remote_id",
					Email:            "some_email@example.com",
				},
			}, nil).AnyTimes()

			mockRemote.EXPECT().MakeSuperuserClient(context.Background()).Return(mockClient)

			mockClient.EXPECT().GetSchedule("some_remote_id", []string{"some_email@example.com"}, gomock.Any(), gomock.Any(), 15).Return([]*remote.ScheduleInformation{tc.sched}, nil)

			mockPluginAPI.EXPECT().GetMattermostUserStatusesByIds([]string{"some_mm_id"}).Return([]*model.Status{&model.Status{Status: tc.currentStatus, UserId: "some_mm_id"}}, nil)

			if tc.newStatus == "" {
				mockPluginAPI.EXPECT().UpdateMattermostUserStatus("some_mm_id", gomock.Any()).Times(0)
			} else {
				mockPluginAPI.EXPECT().UpdateMattermostUserStatus("some_mm_id", tc.newStatus).Times(1)
			}

			a := New(apiConfig, "")
			a.SyncStatusForAllUsers()
		})
	}
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/mock_plugin_api"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote/mock_remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store/mock_store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot/mock_bot"
)

func TestSyncStatusForAllUsers(t *testing.T) {
	for name, tc := range map[string]struct {
		sched         *remote.ScheduleInformation
		currentStatus string
		newStatus     string
	}{
		"User is free but dnd, mark user as online": {
			sched: &remote.ScheduleInformation{
				ScheduleID:       "user_email@example.com",
				AvailabilityView: "0",
			},
			currentStatus: "dnd",
			newStatus:     "online",
		},
		"User is busy but online, mark as dnd": {
			sched: &remote.ScheduleInformation{
				ScheduleID:       "user_email@example.com",
				AvailabilityView: "2",
			},
			currentStatus: "online",
			newStatus:     "dnd",
		},
		"User is free and online, do not change status": {
			sched: &remote.ScheduleInformation{
				ScheduleID:       "user_email@example.com",
				AvailabilityView: "0",
			},
			currentStatus: "online",
			newStatus:     "",
		},
		"User is busy and dnd, do not change status": {
			sched: &remote.ScheduleInformation{
				ScheduleID:       "user_email@example.com",
				AvailabilityView: "2",
			},
			currentStatus: "dnd",
			newStatus:     "",
		},
	} {
		t.Run(name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			s := mock_store.NewMockStore(storeCtrl)

			conf := &config.Config{BotUserID: "bot_mm_id"}

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
			mockPluginAPI := mock_plugin_api.NewMockPluginAPI(pluginAPICtrl)

			env := Env{
				Config: conf,
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
				&store.UserShort{
					MattermostUserID: "bot_mm_id",
					RemoteID:         "bot_remote_id",
					Email:            "bot_email@example.com",
				},
			}, nil).AnyTimes()

			token := &oauth2.Token{
				AccessToken: "bot_oauth_token",
			}
			s.EXPECT().LoadUser("bot_mm_id").Return(&store.User{
				MattermostUserID: "bot_mm_id",
				OAuth2Token:      token,
				Remote: &remote.User{
					ID:   "bot_remote_id",
					Mail: "bot_email@example.com",
				},
			}, nil).Times(2)

			mockRemote.EXPECT().MakeClient(context.Background(), token).Return(mockClient)
			mockClient.EXPECT().GetSuperuserToken().Return("bot_bearer_token", nil)
			mockRemote.EXPECT().MakeSuperuserClient(context.Background(), "bot_bearer_token").Return(mockClient)

			mockClient.EXPECT().GetSchedule("bot_remote_id", []string{"user_email@example.com"}, gomock.Any(), gomock.Any(), 15).Return([]*remote.ScheduleInformation{tc.sched}, nil)

			mockPluginAPI.EXPECT().GetMattermostUserStatusesByIds([]string{"user_mm_id"}).Return([]*model.Status{&model.Status{Status: tc.currentStatus, UserId: "user_mm_id"}}, nil)

			if tc.newStatus == "" {
				mockPluginAPI.EXPECT().UpdateMattermostUserStatus("user_mm_id", gomock.Any()).Times(0)
			} else {
				mockPluginAPI.EXPECT().UpdateMattermostUserStatus("user_mm_id", tc.newStatus).Times(1)
			}

			mscalendar := New(env, "")
			mscalendar.SyncStatusAll()
		})
	}
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

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

func TestSyncStatusAll(t *testing.T) {
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := mock_store.NewMockStore(ctrl)
			poster := mock_bot.NewMockPoster(ctrl)
			mockRemote := mock_remote.NewMockRemote(ctrl)
			mockClient := mock_remote.NewMockClient(ctrl)
			mockPluginAPI := mock_plugin_api.NewMockPluginAPI(ctrl)

			logger := &bot.TestLogger{TB: t}
			conf := &config.Config{BotUserID: "bot_mm_id"}
			env := Env{
				Config: conf,
				Dependencies: &Dependencies{
					Store:             s,
					Logger:            logger,
					Poster:            poster,
					Remote:            mockRemote,
					PluginAPI:         mockPluginAPI,
					IsAuthorizedAdmin: func(mattermostUserID string) (bool, error) { return true, nil },
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

			loc, err := time.LoadLocation("EST")
			require.Nil(t, err)
			hour, minute := 9, 0
			moment := makeTime(hour, minute, loc)
			timeNowFunc = func() time.Time { return moment }

			mockClient.EXPECT().GetSchedule(gomock.Any(), gomock.Any(), gomock.Any(), 15).DoAndReturn(
				func(params []*remote.ScheduleUserInfo, start, end *remote.DateTime, window int) ([]*remote.ScheduleInformation, error) {
					require.Equal(t, "user_email@example.com", params[0].Mail)
					require.Equal(t, "user_remote_id", params[0].RemoteUserID)
					require.Equal(t, "user_remote_id", params[0].RemoteUserID)
					require.Equal(t, time.Duration(0), start.Time().Sub(moment))
					require.Equal(t, time.Duration(15*time.Minute), end.Time().Sub(moment))
					return []*remote.ScheduleInformation{tc.sched}, nil
				})

			mockPluginAPI.EXPECT().GetMattermostUserStatusesByIds([]string{"user_mm_id"}).Return([]*model.Status{{Status: tc.currentStatus, UserId: "user_mm_id"}}, nil)

			if tc.newStatus == "" {
				mockPluginAPI.EXPECT().UpdateMattermostUserStatus("user_mm_id", gomock.Any()).Times(0)
			} else {
				mockPluginAPI.EXPECT().UpdateMattermostUserStatus("user_mm_id", tc.newStatus).Times(1)
			}

			mscalendar := New(env, "")
			res, err := mscalendar.SyncStatusAll()
			require.Nil(t, err)
			require.NotEmpty(t, res)
		})
	}
}

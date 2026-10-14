// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package command

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine/mock_engine"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine/views"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/store"
)

func TestViewCalendar(t *testing.T) {
	testcases := []struct {
		name       string
		setup      func(engine.Engine)
		assertions func(t *testing.T, output string, isUserErr bool, err error)
	}{
		{
			name: "successful view with 24-hour military time",
			setup: func(m engine.Engine) {
				mscal := m.(*mock_engine.MockEngine)
				mscal.EXPECT().GetTimezone(gomock.Any()).Return("UTC", nil).Times(1)
				start := time.Date(2026, 10, 11, 14, 30, 0, 0, time.UTC)
				end := time.Date(2026, 10, 11, 15, 30, 0, 0, time.UTC)
				events := []*remote.Event{
					{
						ID:      "event1",
						Subject: "Team Sync",
						Start:   remote.NewDateTime(start, "UTC"),
						End:     remote.NewDateTime(end, "UTC"),
					},
				}
				mscal.EXPECT().ViewCalendar(gomock.Any(), gomock.Any(), gomock.Any()).Return(events, nil).Times(1)
				mscal.EXPECT().GetUserTimeFormat("mockUserID").Return(views.TimeFormat24Hour).Times(1)
			},
			assertions: func(t *testing.T, output string, isUserErr bool, err error) {
				require.NoError(t, err)
				require.False(t, isUserErr)
				require.Contains(t, output, "Team Sync")
				require.Contains(t, output, "14:30")
				require.Contains(t, output, "15:30")
			},
		},
		{
			name: "successful view with 12-hour time format",
			setup: func(m engine.Engine) {
				mscal := m.(*mock_engine.MockEngine)
				mscal.EXPECT().GetTimezone(gomock.Any()).Return("UTC", nil).Times(1)
				start := time.Date(2026, 10, 11, 14, 30, 0, 0, time.UTC)
				end := time.Date(2026, 10, 11, 15, 30, 0, 0, time.UTC)
				events := []*remote.Event{
					{
						ID:      "event1",
						Subject: "Team Sync",
						Start:   remote.NewDateTime(start, "UTC"),
						End:     remote.NewDateTime(end, "UTC"),
					},
				}
				mscal.EXPECT().ViewCalendar(gomock.Any(), gomock.Any(), gomock.Any()).Return(events, nil).Times(1)
				mscal.EXPECT().GetUserTimeFormat("mockUserID").Return(views.TimeFormat12Hour).Times(1)
			},
			assertions: func(t *testing.T, output string, isUserErr bool, err error) {
				require.NoError(t, err)
				require.False(t, isUserErr)
				require.Contains(t, output, "Team Sync")
				require.Contains(t, output, "2:30 PM")
			},
		},
		{
			name: "error getting timezone",
			setup: func(m engine.Engine) {
				mscal := m.(*mock_engine.MockEngine)
				mscal.EXPECT().GetTimezone(gomock.Any()).Return("", fmt.Errorf("tz error")).Times(1)
			},
			assertions: func(t *testing.T, output string, isUserErr bool, err error) {
				require.Error(t, err)
				require.False(t, isUserErr)
				require.Equal(t, "Error: No timezone found", output)
			},
		},
		{
			name: "user inactive on timezone lookup",
			setup: func(m engine.Engine) {
				mscal := m.(*mock_engine.MockEngine)
				mscal.EXPECT().GetTimezone(gomock.Any()).Return("", fmt.Errorf("%s", store.ErrorUserInactive)).Times(1)
			},
			assertions: func(t *testing.T, output string, isUserErr bool, err error) {
				require.NoError(t, err)
				require.False(t, isUserErr)
				require.Equal(t, store.ErrorUserInactive, output)
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			conf := &config.Config{
				PluginURL: "http://localhost",
			}

			mscal := mock_engine.NewMockEngine(ctrl)
			command := Command{
				Context: &plugin.Context{},
				Args: &model.CommandArgs{
					Command: fmt.Sprintf("/%s viewcal", config.Provider.CommandTrigger),
					UserId:  "mockUserID",
				},
				ChannelID: "mockChannelID",
				Config:    conf,
				Engine:    mscal,
			}

			tt.setup(mscal)

			out, isUserErr, err := command.viewCalendar()
			tt.assertions(t, out, isUserErr, err)
		})
	}
}

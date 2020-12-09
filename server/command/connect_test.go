package command

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/mock_mscalendar"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

func TestConnect(t *testing.T) {
	tcs := []struct {
		name           string
		command        string
		setup          func(m mscalendar.MSCalendar)
		expectedOutput string
		expectedError  string
	}{
		{
			name:    "user already connected",
			command: "connect",
			setup: func(m mscalendar.MSCalendar) {
				mscal := m.(*mock_mscalendar.MockMSCalendar)
				mscal.EXPECT().GetRemoteUser("user_id").Return(&remote.User{Mail: "user@email.com"}, nil).Times(1)
			},
			expectedOutput: "Your Mattermost account is already connected to Microsoft Calendar account `user@email.com`. To connect to a different account, first run `/mscalendar disconnect`.",
			expectedError:  "",
		},
		{
			name:    "user not connected",
			command: "connect",
			setup: func(m mscalendar.MSCalendar) {
				mscal := m.(*mock_mscalendar.MockMSCalendar)
				mscal.EXPECT().GetRemoteUser("user_id").Return(nil, errors.New("remote user not found")).Times(1)
				mscal.EXPECT().Welcome("user_id").Return(nil)
			},
			expectedOutput: "",
			expectedError:  "",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			conf := &config.Config{
				PluginURL: "http://localhost",
			}

			mscal := mock_mscalendar.NewMockMSCalendar(ctrl)
			command := Command{
				Context: &plugin.Context{},
				Args: &model.CommandArgs{
					Command: "/mscalendar " + tc.command,
					UserId:  "user_id",
				},
				ChannelID:  "channel_id",
				Config:     conf,
				MSCalendar: mscal,
			}

			if tc.setup != nil {
				tc.setup(mscal)
			}

			out, _, err := command.Handle()
			if tc.expectedOutput != "" {
				require.Equal(t, tc.expectedOutput, out)
			}

			if tc.expectedError != "" {
				require.Equal(t, tc.expectedError, err.Error())
			} else {
				require.Nil(t, err)
			}
		})
	}
}

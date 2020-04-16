package command

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/mock_mscalendar"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestDisconnect(t *testing.T) {
	tcs := []struct {
		name           string
		command        string
		setup          func(mscalendar.MSCalendar)
		expectedOutput string
		expectedError  string
	}{
		{
			name:    "disconnect failed",
			command: "disconnect",
			setup: func(m mscalendar.MSCalendar) {
				mscal := m.(*mock_mscalendar.MockMSCalendar)
				mscal.EXPECT().DisconnectUser("user_id").Return(errors.New("Some error")).Times(1)
			},
			expectedOutput: "",
			expectedError:  "Command /mscalendar disconnect failed: Some error",
		},
		{
			name:    "disconnect successful",
			command: "disconnect",
			setup: func(m mscalendar.MSCalendar) {
				mscal := m.(*mock_mscalendar.MockMSCalendar)
				mscal.EXPECT().DisconnectUser("user_id").Return(nil).Times(1)
				mscal.EXPECT().ClearSettingsPosts("user_id").Return().Times(1)
			},
			expectedOutput: "Successfully disconnected your account",
			expectedError:  "",
		},
		{
			name:    "non-admin disconnecting bot account",
			command: "disconnect_bot",
			setup: func(m mscalendar.MSCalendar) {
				mscal := m.(*mock_mscalendar.MockMSCalendar)
				mscal.EXPECT().IsAuthorizedAdmin("user_id").Return(false, nil).Times(1)
			},
			expectedOutput: "",
			expectedError:  "Command /mscalendar disconnect_bot failed: non-admin user attempting to disconnect bot account",
		},
		{
			name:    "bot disconnect failed",
			command: "disconnect_bot",
			setup: func(m mscalendar.MSCalendar) {
				mscal := m.(*mock_mscalendar.MockMSCalendar)
				mscal.EXPECT().IsAuthorizedAdmin("user_id").Return(true, nil).Times(1)
				mscal.EXPECT().DisconnectUser("bot_user_id").Return(errors.New("Some error")).Times(1)
			},
			expectedOutput: "",
			expectedError:  "Command /mscalendar disconnect_bot failed: Some error",
		},
		{
			name:    "bot disconnect successful",
			command: "disconnect_bot",
			setup: func(m mscalendar.MSCalendar) {
				mscal := m.(*mock_mscalendar.MockMSCalendar)
				mscal.EXPECT().IsAuthorizedAdmin("user_id").Return(true, nil).Times(1)
				mscal.EXPECT().DisconnectUser("bot_user_id").Return(nil).Times(1)
			},
			expectedOutput: "Successfully disconnected bot user",
			expectedError:  "",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			conf := &config.Config{
				PluginURL: "http://localhost",
				BotUserID: "bot_user_id",
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

			out, err := command.Handle()
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

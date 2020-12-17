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
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
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
			name:    "user not connected",
			command: "disconnect",
			setup: func(m mscalendar.MSCalendar) {
				mscal := m.(*mock_mscalendar.MockMSCalendar)
				mscal.EXPECT().GetRemoteUser("user_id").Return(&remote.User{}, store.ErrNotFound).Times(1)
			},
			expectedOutput: getNotConnectedText(),
			expectedError:  "",
		},
		{
			name:    "error fetching user",
			command: "disconnect",
			setup: func(m mscalendar.MSCalendar) {
				mscal := m.(*mock_mscalendar.MockMSCalendar)
				mscal.EXPECT().GetRemoteUser("user_id").Return(&remote.User{}, errors.New("some error")).Times(1)
			},
			expectedOutput: "",
			expectedError:  "Command /mscalendar disconnect failed: some error",
		},
		{
			name:    "disconnect failed",
			command: "disconnect",
			setup: func(m mscalendar.MSCalendar) {
				mscal := m.(*mock_mscalendar.MockMSCalendar)
				mscal.EXPECT().GetRemoteUser("user_id").Return(&remote.User{}, nil).Times(1)
				mscal.EXPECT().DisconnectUser("user_id").Return(errors.New("some error")).Times(1)
			},
			expectedOutput: "",
			expectedError:  "Command /mscalendar disconnect failed: some error",
		},
		{
			name:    "disconnect successful",
			command: "disconnect",
			setup: func(m mscalendar.MSCalendar) {
				mscal := m.(*mock_mscalendar.MockMSCalendar)
				mscal.EXPECT().GetRemoteUser("user_id").Return(&remote.User{}, nil).Times(1)
				mscal.EXPECT().DisconnectUser("user_id").Return(nil).Times(1)
				mscal.EXPECT().ClearSettingsPosts("user_id").Return().Times(1)
			},
			expectedOutput: "Successfully disconnected your account",
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

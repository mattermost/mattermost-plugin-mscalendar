package command

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/api"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/api/mock_api"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestConnect(t *testing.T) {
	tcs := []struct {
		name           string
		command        string
		setup          func(api.API)
		expectedOutput string
		expectedError  string
	}{
		{
			name:    "user already connected",
			command: "connect",
			setup: func(a api.API) {
				mapi := a.(*mock_api.MockAPI)
				mapi.EXPECT().GetRemoteUser("user_id").Return(&remote.User{}, nil).Times(1)
			},
			expectedOutput: "Your account is already connected. Please run `/mscalendar disconnect`",
			expectedError:  "",
		},
		{
			name:    "user not connected",
			command: "connect",
			setup: func(a api.API) {
				mapi := a.(*mock_api.MockAPI)
				mapi.EXPECT().GetRemoteUser("user_id").Return(nil, errors.New("remote user not found")).Times(1)
			},
			expectedOutput: "[Click here to link your Microsoft Calendar account.](http://localhost/oauth2/connect)",
			expectedError:  "",
		},
		{
			name:    "non-admin connecting bot account",
			command: "connect_bot",
			setup: func(a api.API) {
				mapi := a.(*mock_api.MockAPI)
				mapi.EXPECT().IsAuthorizedAdmin("user_id").Return(false, nil).Times(1)
			},
			expectedOutput: "",
			expectedError:  "Command /mscalendar connect_bot failed: non-admin user attempting to connect bot account",
		},
		{
			name:    "bot user already connected",
			command: "connect_bot",
			setup: func(a api.API) {
				mapi := a.(*mock_api.MockAPI)
				mapi.EXPECT().IsAuthorizedAdmin("user_id").Return(true, nil).Times(1)
				mapi.EXPECT().GetRemoteUser("bot_user_id").Return(&remote.User{}, nil).Times(1)
			},
			expectedOutput: "Bot user already connected. Please run `/mscalendar disconnect_bot`",
			expectedError:  "",
		},
		{
			name:    "bot user not connected",
			command: "connect_bot",
			setup: func(a api.API) {
				mapi := a.(*mock_api.MockAPI)
				mapi.EXPECT().IsAuthorizedAdmin("user_id").Return(true, nil).Times(1)
				mapi.EXPECT().GetRemoteUser("bot_user_id").Return(nil, errors.New("remote user not found")).Times(1)
			},
			expectedOutput: "[Click here to link the bot's Microsoft Calendar account.](http://localhost/oauth2/connect?bot=true)",
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

			mapi := mock_api.NewMockAPI(ctrl)
			command := Command{
				Context: &plugin.Context{},
				Args: &model.CommandArgs{
					Command: "/mscalendar " + tc.command,
					UserId:  "user_id",
				},
				ChannelID: "channel_id",
				Config:    conf,
				API:       mapi,
			}

			if tc.setup != nil {
				tc.setup(mapi)
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

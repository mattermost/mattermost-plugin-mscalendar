// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/mock_plugin_api"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote/mock_remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store/mock_store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot/mock_bot"
)

func newTestMattermostChannel() *model.Channel {
	return &model.Channel{
		Id:   "mattermost_channel_id",
		Type: model.CHANNEL_DIRECT,
	}
}

func newTestStoreUser(activeEvents []string, mattermostUserID string, autoRespond bool, autoRespondMessage string) *store.User {
	return &store.User{
		ActiveEvents:     activeEvents,
		MattermostUserID: mattermostUserID,
		Settings: store.Settings{
			AutoRespond:        autoRespond,
			AutoRespondMessage: autoRespondMessage,
		},
	}
}

func newTestMattermostStatus(status string) *model.Status {
	return &model.Status{
		Status: status,
	}
}

func newTestMattermostUser(userID string) *model.User {
	return &model.User{
		Id: userID,
	}
}

func newTestMattermostPost() *model.Post {
	return &model.Post{
		ChannelId: "mattermost_post_channel_id",
		UserId:    "mattermost_user_sender_id",
	}
}

func TestHandleBusyDM(t *testing.T) {
	tcs := []struct {
		name                      string
		expectedError             string
		recipientActiveEvents     []string
		recipientStatusString     string
		autoRespondSetting        bool
		autoRespondMessageSetting string
		expectedDMMessage         string
	}{
		{
			name:                      "Happy path, bot responds to DM",
			expectedError:             "",
			recipientActiveEvents:     []string{"active_event_hash"},
			recipientStatusString:     model.STATUS_DND,
			autoRespondSetting:        true,
			autoRespondMessageSetting: "Hello, I'm in a meeting and will respond to your message as soon as I'm free.",
			expectedDMMessage:         "Hello, I'm in a meeting and will respond to your message as soon as I'm free.",
		}, {
			name:                      "Auto-respond message not set, fall back to default",
			expectedError:             "",
			recipientActiveEvents:     []string{"active_event_hash"},
			recipientStatusString:     model.STATUS_DND,
			autoRespondSetting:        true,
			autoRespondMessageSetting: "",
			expectedDMMessage:         "This user is currently in a meeting.",
		}, {
			name:                      "Recipient has no active events",
			expectedError:             "",
			recipientActiveEvents:     []string{},
			recipientStatusString:     model.STATUS_DND,
			autoRespondSetting:        true,
			autoRespondMessageSetting: "Hello, I'm in a meeting and will respond to your message as soon as I'm free.",
			expectedDMMessage:         "",
		}, {
			name:                      "Recipient autorespond Setting turned off",
			expectedError:             "",
			recipientActiveEvents:     []string{"active_event_hash"},
			recipientStatusString:     model.STATUS_DND,
			autoRespondSetting:        false,
			autoRespondMessageSetting: "Hello, I'm in a meeting and will respond to your message as soon as I'm free.",
			expectedDMMessage:         "",
		}, {
			name:                      "Recipient user status is set to online",
			expectedError:             "",
			recipientActiveEvents:     []string{"active_event_hash"},
			recipientStatusString:     model.STATUS_ONLINE,
			autoRespondSetting:        true,
			autoRespondMessageSetting: "Hello, I'm in a meeting and will respond to your message as soon as I'm free.",
			expectedDMMessage:         "",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mock_store.NewMockStore(ctrl)
			mockPoster := mock_bot.NewMockPoster(ctrl)
			mockRemote := mock_remote.NewMockRemote(ctrl)
			mockPluginAPI := mock_plugin_api.NewMockPluginAPI(ctrl)

			conf := &config.Config{PluginVersion: "x.x.x"}
			env := Env{
				Config: conf,
				Dependencies: &Dependencies{
					Store:     mockStore,
					Logger:    &bot.NilLogger{},
					Poster:    mockPoster,
					Remote:    mockRemote,
					PluginAPI: mockPluginAPI,
				},
			}

			post := newTestMattermostPost()
			channel := newTestMattermostChannel()
			mattermostUserSender := newTestMattermostUser("mattermost_user_sender_id")
			mattermostUserRecipient := newTestMattermostUser("mattermost_user_recipient_id")

			recipientStatus := newTestMattermostStatus(tc.recipientStatusString)
			storedRecipient := newTestStoreUser(tc.recipientActiveEvents, "mattermost_user_recipient_id", tc.autoRespondSetting, tc.autoRespondMessageSetting)

			mockPluginAPI.EXPECT().GetMattermostChannel("mattermost_post_channel_id").Return(channel, nil)
			mockPluginAPI.EXPECT().GetMattermostUsersInChannel("mattermost_post_channel_id", model.CHANNEL_SORT_BY_USERNAME, 0, 2).Return([]*model.User{mattermostUserSender, mattermostUserRecipient}, nil)

			mockStore.EXPECT().LoadUser("mattermost_user_sender_id").Return(nil, errors.New("user not found"))
			mockStore.EXPECT().LoadUser("mattermost_user_recipient_id").Return(storedRecipient, nil)

			if tc.autoRespondSetting && len(tc.recipientActiveEvents) > 0 {
				mockPluginAPI.EXPECT().GetMattermostUserStatus("mattermost_user_recipient_id").Return(recipientStatus, nil)
			}

			if tc.expectedDMMessage != "" {
				mockPoster.EXPECT().Ephemeral("mattermost_user_sender_id", "mattermost_post_channel_id", tc.expectedDMMessage)
			}

			m := New(env, post.UserId)
			err := m.HandleBusyDM(post)

			if tc.expectedError != "" {
				require.Equal(t, tc.expectedError, err.Error())
			} else {
				require.Nil(t, err)
			}
		})
	}
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package engine

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot/mock_bot"
)


func TestNewConnectAttachment(t *testing.T) {
	env := Env{
		Config: &config.Config{
			Provider: config.ProviderConfig{
				DisplayName:    "Microsoft Calendar",
				CommandTrigger: "mscalendar",
			},
		},
		Dependencies: &Dependencies{},
	}
	pluginURL := "http://localhost:8065/plugins/com.mattermost.mscalendar"

	b := &mscBot{
		Env:       env,
		pluginURL: pluginURL,
	}

	attachment := b.newConnectAttachment()

	require.NotNil(t, attachment)
	require.Equal(t, "Connect your Microsoft Calendar account", attachment.Title)
	require.Equal(t, ConnectMessageBody, attachment.Text)
	require.Len(t, attachment.Actions, 1)

	action := attachment.Actions[0]
	require.Equal(t, ConnectActionID, action.Id)
	require.Equal(t, "Connect account", action.Name)
	require.Equal(t, model.PostActionTypeButton, action.Type)
	require.NotNil(t, action.Integration)
	require.Equal(t, pluginURL+"/oauth2/connect", action.Integration.URL)

	expectedFallback := fmt.Sprintf("Connect your Microsoft Calendar account\n%s\n[Connect account](%s/oauth2/connect)", ConnectMessageBody, pluginURL)
	require.Equal(t, expectedFallback, attachment.Fallback)
}

func TestNewConnectedAttachment(t *testing.T) {
	env := Env{
		Config: &config.Config{
			Provider: config.ProviderConfig{
				DisplayName:    "Microsoft Calendar",
				CommandTrigger: "mscalendar",
			},
		},
		Dependencies: &Dependencies{},
	}

	b := &mscBot{
		Env:       env,
		pluginURL: "http://localhost:8065/plugins/com.mattermost.mscalendar",
	}

	attachment := b.newConnectedAttachment("user@example.com")

	require.NotNil(t, attachment)
	require.Equal(t, "Connect", attachment.Title)
	require.Equal(t, ":tada: Congratulations! Your Microsoft Calendar account (*user@example.com*) has been connected to Mattermost.", attachment.Text)
}

type testBotPosterWrapper struct {
	bot.Bot
	poster *mock_bot.MockPoster
}

func (w *testBotPosterWrapper) DM(userID, format string, args ...interface{}) (string, error) {
	return w.poster.DM(userID, format, args...)
}

func TestNotifySettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPoster := mock_bot.NewMockPoster(ctrl)
	config.Provider = config.ProviderConfig{
		DisplayName:    "Microsoft Calendar",
		CommandTrigger: "mscalendar",
	}

	b := &mscBot{
		Bot: &testBotPosterWrapper{poster: mockPoster},
		Env: Env{
			Config: &config.Config{
				Provider: config.Provider,
			},
		},
	}

	mockPoster.EXPECT().DM("test_user_id", "Feel free to change these settings anytime by typing `/%s settings`", "mscalendar").Return("post_id", nil).Times(1)

	err := b.notifySettings("test_user_id")
	require.NoError(t, err)
}


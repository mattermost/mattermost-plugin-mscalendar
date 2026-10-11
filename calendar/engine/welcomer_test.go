// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package engine

import (
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine/mock_plugin_api"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/store/mock_store"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/tracker/mock_tracker"
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

type testWelcomeBotWrapper struct {
	bot.Bot
	dmWithAttachments func(userID string, attachments ...*model.MessageAttachment) (string, error)
	deletePost        func(postID string) error
	updatePost        func(post *model.Post) error
	startFlow         func(userID string) error
	cancelFlow        func(userID string) error
	errorf            func(format string, args ...interface{})
}

func (w *testWelcomeBotWrapper) DMWithAttachments(userID string, attachments ...*model.MessageAttachment) (string, error) {
	if w.dmWithAttachments != nil {
		return w.dmWithAttachments(userID, attachments...)
	}
	return "", nil
}

func (w *testWelcomeBotWrapper) DeletePost(postID string) error {
	if w.deletePost != nil {
		return w.deletePost(postID)
	}
	return nil
}

func (w *testWelcomeBotWrapper) UpdatePost(post *model.Post) error {
	if w.updatePost != nil {
		return w.updatePost(post)
	}
	return nil
}

func (w *testWelcomeBotWrapper) Start(userID string) error {
	if w.startFlow != nil {
		return w.startFlow(userID)
	}
	return nil
}

func (w *testWelcomeBotWrapper) Cancel(userID string) error {
	if w.cancelFlow != nil {
		return w.cancelFlow(userID)
	}
	return nil
}

func (w *testWelcomeBotWrapper) Errorf(format string, args ...interface{}) {
	if w.errorf != nil {
		w.errorf(format, args...)
	}
}


func TestWelcome_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_store.NewMockStore(ctrl)
	postDeleted := false
	dmCalled := false

	testBot := &testWelcomeBotWrapper{
		deletePost: func(postID string) error {
			if postID == "old_post" {
				postDeleted = true
			}
			return nil
		},
		dmWithAttachments: func(userID string, attachments ...*model.MessageAttachment) (string, error) {
			dmCalled = true
			require.Equal(t, "user_123", userID)
			require.Len(t, attachments, 1)
			require.Equal(t, "Connect your Microsoft Calendar account", attachments[0].Title)
			return "new_post_id", nil
		},
	}

	b := &mscBot{
		Bot: testBot,
		Env: Env{
			Config: &config.Config{
				Provider: config.ProviderConfig{
					DisplayName:    "Microsoft Calendar",
					CommandTrigger: "mscalendar",
				},
			},
			Dependencies: &Dependencies{
				Store: mockStore,
			},
		},
		pluginURL: "http://localhost:8065/plugins/com.mattermost.mscalendar",
	}

	mockStore.EXPECT().DeleteUserWelcomePost("user_123").Return("old_post", nil).Times(1)
	mockStore.EXPECT().StoreUserWelcomePost("user_123", "new_post_id").Return(nil).Times(1)

	err := b.Welcome("user_123")
	require.NoError(t, err)
	require.True(t, postDeleted)
	require.True(t, dmCalled)
}

func TestWelcome_DMError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_store.NewMockStore(ctrl)

	testBot := &testWelcomeBotWrapper{
		dmWithAttachments: func(userID string, attachments ...*model.MessageAttachment) (string, error) {
			return "", errors.New("DM failure")
		},
	}

	b := &mscBot{
		Bot: testBot,
		Env: Env{
			Config: &config.Config{
				Provider: config.ProviderConfig{
					DisplayName:    "Microsoft Calendar",
					CommandTrigger: "mscalendar",
				},
			},
			Dependencies: &Dependencies{
				Store: mockStore,
			},
		},
		pluginURL: "http://localhost:8065/plugins/com.mattermost.mscalendar",
	}

	mockStore.EXPECT().DeleteUserWelcomePost("user_123").Return("", nil).Times(1)

	err := b.Welcome("user_123")
	require.Error(t, err)
	require.Contains(t, err.Error(), "DM failure")
}

func TestAfterSuccessfullyConnect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPluginAPI := mock_plugin_api.NewMockPluginAPI(ctrl)
	mockTracker := mock_tracker.NewMockTracker(ctrl)
	mockStore := mock_store.NewMockStore(ctrl)

	postUpdated := false
	flowStarted := false

	testBot := &testWelcomeBotWrapper{
		updatePost: func(post *model.Post) error {
			if post.Id == "welcome_post_1" {
				postUpdated = true
			}
			return nil
		},
		startFlow: func(userID string) error {
			if userID == "user_123" {
				flowStarted = true
			}
			return nil
		},
	}

	b := &mscBot{
		Bot: testBot,
		Env: Env{
			Config: &config.Config{
				Provider: config.ProviderConfig{
					DisplayName:    "Microsoft Calendar",
					CommandTrigger: "mscalendar",
				},
			},
			Dependencies: &Dependencies{
				PluginAPI: mockPluginAPI,
				Tracker:   mockTracker,
				Store:     mockStore,
			},
		},
	}

	mockPluginAPI.EXPECT().PublishWebsocketEvent("user_123", "connected", map[string]any{"action": "connected"}).Times(1)
	mockTracker.EXPECT().TrackUserAuthenticated("user_123").Times(1)
	mockStore.EXPECT().DeleteUserWelcomePost("user_123").Return("welcome_post_1", nil).Times(1)

	err := b.AfterSuccessfullyConnect("user_123", "user@example.com")
	require.NoError(t, err)
	require.True(t, postUpdated)
	require.True(t, flowStarted)
}

func TestAfterDisconnect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPluginAPI := mock_plugin_api.NewMockPluginAPI(ctrl)
	mockTracker := mock_tracker.NewMockTracker(ctrl)
	mockStore := mock_store.NewMockStore(ctrl)

	flowCancelled := false
	postDeleted := false

	testBot := &testWelcomeBotWrapper{
		cancelFlow: func(userID string) error {
			if userID == "user_123" {
				flowCancelled = true
			}
			return nil
		},
		deletePost: func(postID string) error {
			if postID == "welcome_post_1" {
				postDeleted = true
			}
			return nil
		},
	}

	b := &mscBot{
		Bot: testBot,
		Env: Env{
			Dependencies: &Dependencies{
				PluginAPI: mockPluginAPI,
				Tracker:   mockTracker,
				Store:     mockStore,
			},
		},
	}

	mockPluginAPI.EXPECT().PublishWebsocketEvent("user_123", "disconnected", map[string]any{"action": "disconnected"}).Times(1)
	mockTracker.EXPECT().TrackUserDeauthenticated("user_123").Times(1)
	mockStore.EXPECT().DeleteUserWelcomePost("user_123").Return("welcome_post_1", nil).Times(1)

	err := b.AfterDisconnect("user_123")
	require.NoError(t, err)
	require.True(t, flowCancelled)
	require.True(t, postDeleted)
}




// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package jobs

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine/mock_plugin_api"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote/mock_remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/store/mock_store"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot"
)

// noopLogger is a bot.Logger that discards everything, so tests can focus on
// the data dependencies without matching variadic log calls.
type noopLogger struct{}

func (noopLogger) With(bot.LogContext) bot.Logger { return noopLogger{} }
func (noopLogger) Timed() bot.Logger              { return noopLogger{} }
func (noopLogger) Debugf(string, ...interface{})  {}
func (noopLogger) Errorf(string, ...interface{})  {}
func (noopLogger) Infof(string, ...interface{})   {}
func (noopLogger) Warnf(string, ...interface{})   {}

func TestRunRenewJob(t *testing.T) {
	const (
		connectedUserID = "connectedUserID"
		connectedRemote = "connectedRemoteID"
	)

	tests := []struct {
		name  string
		setup func(*mock_store.MockStore, *mock_remote.MockRemote, *mock_plugin_api.MockPluginAPI)
	}{
		{
			name: "skips nil index entries and users without a connected remote account",
			setup: func(mockStore *mock_store.MockStore, _ *mock_remote.MockRemote, _ *mock_plugin_api.MockPluginAPI) {
				index := store.UserIndex{
					nil,
					&store.UserShort{MattermostUserID: "unconnectedUserID", RemoteID: ""},
				}
				mockStore.EXPECT().LoadUserIndex().Return(index, nil)
				// No LoadUser/MakeUserClient expectations: a skipped user must not
				// reach the renewal path. gomock fails the test on any extra call.
			},
		},
		{
			name: "processes connected users",
			setup: func(mockStore *mock_store.MockStore, mockRemote *mock_remote.MockRemote, mockPluginAPI *mock_plugin_api.MockPluginAPI) {
				index := store.UserIndex{
					&store.UserShort{MattermostUserID: connectedUserID, RemoteID: connectedRemote},
				}
				mockStore.EXPECT().LoadUserIndex().Return(index, nil)

				// A connected user must reach the renewal path. We stop it early by
				// failing client creation, which exercises the loop's error handling
				// without standing up the full subscription chain.
				mockStore.EXPECT().LoadUser(connectedUserID).Return(&store.User{
					Remote: &remote.User{ID: connectedRemote},
				}, nil)
				mockPluginAPI.EXPECT().GetMattermostUser(connectedUserID).Return(&model.User{Id: connectedUserID}, nil)
				mockRemote.EXPECT().MakeUserClient(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("client creation failed"))
			},
		},
		{
			name: "returns early when the user index fails to load",
			setup: func(mockStore *mock_store.MockStore, _ *mock_remote.MockRemote, _ *mock_plugin_api.MockPluginAPI) {
				mockStore.EXPECT().LoadUserIndex().Return(nil, errors.New("load failed"))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mock_store.NewMockStore(ctrl)
			mockRemote := mock_remote.NewMockRemote(ctrl)
			mockPluginAPI := mock_plugin_api.NewMockPluginAPI(ctrl)

			tc.setup(mockStore, mockRemote, mockPluginAPI)

			env := engine.Env{
				Config: &config.Config{},
				Dependencies: &engine.Dependencies{
					Store:     mockStore,
					Remote:    mockRemote,
					PluginAPI: mockPluginAPI,
					Logger:    noopLogger{},
				},
			}

			assert.NotPanics(t, func() {
				runRenewJob(env)
			})
		})
	}
}

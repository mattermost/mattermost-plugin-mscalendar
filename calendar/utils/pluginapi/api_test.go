// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package pluginapi

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCanReadChannel(t *testing.T) {
	const (
		channelID = "channelID"
		userID    = "userID"
		teamID    = "teamID"
	)

	tests := []struct {
		name     string
		setup    func(*plugintest.API)
		expected bool
	}{
		{
			name: "channel member is granted read access",
			setup: func(api *plugintest.API) {
				api.On("HasPermissionToChannel", userID, channelID, model.PermissionReadChannelContent).Return(true)
			},
			expected: true,
		},
		{
			name: "non-member of a private channel is denied without team fallback",
			setup: func(api *plugintest.API) {
				api.On("HasPermissionToChannel", userID, channelID, model.PermissionReadChannelContent).Return(false)
				api.On("GetChannel", channelID).Return(&model.Channel{Id: channelID, TeamId: teamID, Type: model.ChannelTypePrivate}, nil)
			},
			expected: false,
		},
		{
			name: "non-member of a DM channel is denied",
			setup: func(api *plugintest.API) {
				api.On("HasPermissionToChannel", userID, channelID, model.PermissionReadChannelContent).Return(false)
				api.On("GetChannel", channelID).Return(&model.Channel{Id: channelID, Type: model.ChannelTypeDirect}, nil)
			},
			expected: false,
		},
		{
			name: "non-member of an open channel falls back to team permission (granted)",
			setup: func(api *plugintest.API) {
				api.On("HasPermissionToChannel", userID, channelID, model.PermissionReadChannelContent).Return(false)
				api.On("GetChannel", channelID).Return(&model.Channel{Id: channelID, TeamId: teamID, Type: model.ChannelTypeOpen}, nil)
				api.On("HasPermissionToTeam", userID, teamID, model.PermissionReadPublicChannel).Return(true)
			},
			expected: true,
		},
		{
			name: "non-member of an open channel without team permission is denied",
			setup: func(api *plugintest.API) {
				api.On("HasPermissionToChannel", userID, channelID, model.PermissionReadChannelContent).Return(false)
				api.On("GetChannel", channelID).Return(&model.Channel{Id: channelID, TeamId: teamID, Type: model.ChannelTypeOpen}, nil)
				api.On("HasPermissionToTeam", userID, teamID, model.PermissionReadPublicChannel).Return(false)
			},
			expected: false,
		},
		{
			name: "channel lookup failure denies access",
			setup: func(api *plugintest.API) {
				api.On("HasPermissionToChannel", userID, channelID, model.PermissionReadChannelContent).Return(false)
				api.On("GetChannel", channelID).Return(nil, &model.AppError{Message: "not found"})
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &plugintest.API{}
			defer mockAPI.AssertExpectations(t)
			tc.setup(mockAPI)

			a := New(mockAPI)
			assert.Equal(t, tc.expected, a.CanReadChannel(channelID, userID))

			mockAPI.AssertNotCalled(t, "HasPermissionToChannel", mock.Anything, mock.Anything, model.PermissionReadChannel)
		})
	}
}

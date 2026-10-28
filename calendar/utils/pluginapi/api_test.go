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

func TestGetViewUsersRestrictions(t *testing.T) {
	const viewerID = "viewerID"
	appErr := &model.AppError{Message: "error"}

	tests := []struct {
		name        string
		setup       func(*plugintest.API)
		expected    *model.ViewUsersRestrictions
		expectedErr bool
	}{
		{
			name: "global permission means no restrictions",
			setup: func(api *plugintest.API) {
				api.On("HasPermissionTo", viewerID, model.PermissionViewMembers).Return(true)
			},
			expected: nil,
		},
		{
			name: "collects permitted teams and every channel page",
			setup: func(api *plugintest.API) {
				api.On("HasPermissionTo", viewerID, model.PermissionViewMembers).Return(false)
				api.On("GetTeamsForUser", viewerID).Return([]*model.Team{{Id: "teamA"}, {Id: "teamB"}}, nil)
				api.On("HasPermissionToTeam", viewerID, "teamA", model.PermissionViewMembers).Return(true)
				api.On("HasPermissionToTeam", viewerID, "teamB", model.PermissionViewMembers).Return(false)
				firstPage := make([]*model.ChannelMember, membershipsPageSize)
				for i := range firstPage {
					firstPage[i] = &model.ChannelMember{ChannelId: "channelA"}
				}
				api.On("GetChannelMembersForUser", "", viewerID, 0, membershipsPageSize).Return(firstPage, nil)
				api.On("GetChannelMembersForUser", "", viewerID, 1, membershipsPageSize).Return([]*model.ChannelMember{{ChannelId: "channelB"}}, nil)
			},
			expected: func() *model.ViewUsersRestrictions {
				channels := make([]string, 0, membershipsPageSize+1)
				for i := 0; i < membershipsPageSize; i++ {
					channels = append(channels, "channelA")
				}
				return &model.ViewUsersRestrictions{Teams: []string{"teamA"}, Channels: append(channels, "channelB")}
			}(),
		},
		{
			name: "teams lookup failure",
			setup: func(api *plugintest.API) {
				api.On("HasPermissionTo", viewerID, model.PermissionViewMembers).Return(false)
				api.On("GetTeamsForUser", viewerID).Return(nil, appErr)
			},
			expectedErr: true,
		},
		{
			name: "channel memberships lookup failure",
			setup: func(api *plugintest.API) {
				api.On("HasPermissionTo", viewerID, model.PermissionViewMembers).Return(false)
				api.On("GetTeamsForUser", viewerID).Return([]*model.Team{}, nil)
				api.On("GetChannelMembersForUser", "", viewerID, 0, membershipsPageSize).Return(nil, appErr)
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &plugintest.API{}
			defer mockAPI.AssertExpectations(t)
			tc.setup(mockAPI)

			result, err := New(mockAPI).GetViewUsersRestrictions(viewerID)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFilterVisibleUsers(t *testing.T) {
	const (
		viewerID  = "viewerID"
		teamID    = "teamID"
		channelA  = "channelA"
		channelB  = "channelB"
		userA     = "userA"
		userB     = "userB"
		otherUser = "otherTeamUser"
	)
	appErr := &model.AppError{Message: "error"}
	channelsOnly := &model.ViewUsersRestrictions{Teams: []string{}, Channels: []string{channelA, channelB}}
	teamOnly := &model.ViewUsersRestrictions{Teams: []string{teamID}, Channels: []string{}}

	tests := []struct {
		name         string
		restrictions *model.ViewUsersRestrictions
		userIDs      []string
		setup        func(*plugintest.API)
		expected     []string
		expectedErr  bool
	}{
		{
			name:         "nil restrictions allow every user",
			restrictions: nil,
			userIDs:      []string{userA, otherUser},
			setup:        func(api *plugintest.API) {},
			expected:     []string{userA, otherUser},
		},
		{
			name:         "viewer is always visible to themselves",
			restrictions: &model.ViewUsersRestrictions{},
			userIDs:      []string{viewerID},
			setup:        func(api *plugintest.API) {},
			expected:     []string{viewerID},
		},
		{
			name:         "active member of a permitted team is visible",
			restrictions: teamOnly,
			userIDs:      []string{userA, userB, otherUser},
			setup: func(api *plugintest.API) {
				api.On("GetTeamMembersForUser", userA, 0, membershipsPageSize).Return([]*model.TeamMember{{TeamId: teamID}}, nil)
				api.On("GetTeamMembersForUser", userB, 0, membershipsPageSize).Return([]*model.TeamMember{{TeamId: teamID, DeleteAt: 1}}, nil)
				api.On("GetTeamMembersForUser", otherUser, 0, membershipsPageSize).Return([]*model.TeamMember{{TeamId: "otherTeam"}}, nil)
			},
			expected: []string{userA},
		},
		{
			name:         "team memberships are paged",
			restrictions: teamOnly,
			userIDs:      []string{userA},
			setup: func(api *plugintest.API) {
				firstPage := make([]*model.TeamMember, membershipsPageSize)
				for i := range firstPage {
					firstPage[i] = &model.TeamMember{TeamId: "otherTeam"}
				}
				api.On("GetTeamMembersForUser", userA, 0, membershipsPageSize).Return(firstPage, nil)
				api.On("GetTeamMembersForUser", userA, 1, membershipsPageSize).Return([]*model.TeamMember{{TeamId: teamID}}, nil)
			},
			expected: []string{userA},
		},
		{
			name:         "channel members are resolved in batches and stop once all are found",
			restrictions: channelsOnly,
			userIDs:      []string{userA, userB, otherUser},
			setup: func(api *plugintest.API) {
				api.On("GetChannelMembersByIds", channelA, []string{userA, userB, otherUser}).Return(model.ChannelMembers{{UserId: userB}}, nil)
				api.On("GetChannelMembersByIds", channelB, []string{userA, otherUser}).Return(model.ChannelMembers{{UserId: userA}, {UserId: otherUser}}, nil)
			},
			expected: []string{userA, userB, otherUser},
		},
		{
			name:         "user sharing no team or channel is hidden",
			restrictions: &model.ViewUsersRestrictions{Channels: []string{channelA}},
			userIDs:      []string{otherUser},
			setup: func(api *plugintest.API) {
				api.On("GetChannelMembersByIds", channelA, []string{otherUser}).Return(model.ChannelMembers{}, nil)
			},
			expected: []string{},
		},
		{
			name:         "team memberships lookup failure",
			restrictions: teamOnly,
			userIDs:      []string{userA},
			setup: func(api *plugintest.API) {
				api.On("GetTeamMembersForUser", userA, 0, membershipsPageSize).Return(nil, appErr)
			},
			expectedErr: true,
		},
		{
			name:         "channel members lookup failure",
			restrictions: &model.ViewUsersRestrictions{Channels: []string{channelA}},
			userIDs:      []string{userA},
			setup: func(api *plugintest.API) {
				api.On("GetChannelMembersByIds", channelA, []string{userA}).Return(nil, appErr)
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &plugintest.API{}
			defer mockAPI.AssertExpectations(t)
			tc.setup(mockAPI)

			result, err := New(mockAPI).FilterVisibleUsers(viewerID, tc.restrictions, tc.userIDs)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

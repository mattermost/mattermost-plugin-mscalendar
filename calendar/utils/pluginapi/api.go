// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package pluginapi

import (
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/store"
)

type API struct {
	api plugin.API
}

func New(api plugin.API) *API {
	return &API{
		api: api,
	}
}

func (a *API) SearchLinkableChannelForUser(teamID, mattermostUserID, search string) ([]*model.Channel, error) {
	channels, err := a.api.SearchChannels(teamID, search)
	if err != nil {
		return nil, err
	}

	var result []*model.Channel
	for _, ch := range channels {
		if a.CanLinkEventToChannel(ch.Id, mattermostUserID) {
			result = append(result, ch)
		}
	}

	return result, nil
}

func (a *API) GetMattermostUserStatus(mattermostUserID string) (*model.Status, error) {
	st, err := a.api.GetUserStatus(mattermostUserID)
	if err != nil {
		return nil, err
	}

	return st, nil
}

func (a *API) GetMattermostUserStatusesByIds(mattermostUserIDs []string) ([]*model.Status, error) {
	st, err := a.api.GetUserStatusesByIds(mattermostUserIDs)
	if err != nil {
		return nil, err
	}

	return st, nil
}

func (a *API) UpdateMattermostUserStatus(mattermostUserID, status string) (*model.Status, error) {
	s, err := a.api.UpdateUserStatus(mattermostUserID, status)
	if err != nil {
		return s, err
	}
	return s, nil
}

func (a *API) UpdateMattermostUserCustomStatus(mattermostUserID string, customStatus *model.CustomStatus) *model.AppError {
	return a.api.UpdateUserCustomStatus(mattermostUserID, customStatus)
}

func (a *API) RemoveMattermostUserCustomStatus(mattermostUserID string) *model.AppError {
	return a.api.RemoveUserCustomStatus(mattermostUserID)
}

// IsSysAdmin returns true if the user is authorized to use the workflow plugin's admin-level APIs/commands.
func (a *API) IsSysAdmin(mattermostUserID string) (bool, error) {
	user, err := a.api.GetUser(mattermostUserID)
	if err != nil {
		return false, err
	}
	return user.IsSystemAdmin(), nil
}

func (a *API) GetMattermostUserByUsername(mattermostUsername string) (*model.User, error) {
	for strings.HasPrefix(mattermostUsername, "@") {
		mattermostUsername = mattermostUsername[1:]
	}
	u, err := a.api.GetUserByUsername(mattermostUsername)
	if err != nil {
		return nil, err
	}
	if u.DeleteAt != 0 {
		return nil, store.ErrNotFound
	}
	return u, nil
}

func (a *API) GetMattermostUser(mattermostUserID string) (*model.User, error) {
	mmuser, err := a.api.GetUser(mattermostUserID)
	if err != nil {
		return nil, err
	}
	if mmuser.DeleteAt != 0 {
		return nil, store.ErrNotFound
	}
	return mmuser, nil
}

func (a *API) GetMattermostUserTeams(mattermostUserID string) ([]*model.Team, error) {
	teams, err := a.api.GetTeamsForUser(mattermostUserID)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (a *API) CanLinkEventToChannel(channelID, userID string) bool {
	return a.api.HasPermissionToChannel(userID, channelID, model.PermissionCreatePost)
}

// CanReadChannel reports whether the user is allowed to read the given channel,
// mirroring the authorization the core post API enforces before returning a post.
func (a *API) CanReadChannel(channelID, userID string) bool {
	if a.api.HasPermissionToChannel(userID, channelID, model.PermissionReadChannelContent) {
		return true
	}

	channel, appErr := a.api.GetChannel(channelID)
	if appErr != nil {
		return false
	}

	if channel.Type == model.ChannelTypeOpen {
		return a.api.HasPermissionToTeam(userID, channel.TeamId, model.PermissionReadPublicChannel)
	}

	return false
}

const membershipsPageSize = 200

// GetViewUsersRestrictions returns the teams and channels that limit which users
// the viewer can see, mirroring core. A nil result means the viewer can see everyone.
func (a *API) GetViewUsersRestrictions(viewerID string) (*model.ViewUsersRestrictions, error) {
	if a.api.HasPermissionTo(viewerID, model.PermissionViewMembers) {
		return nil, nil
	}

	teams, appErr := a.api.GetTeamsForUser(viewerID)
	if appErr != nil {
		return nil, appErr
	}
	restrictions := &model.ViewUsersRestrictions{Teams: []string{}, Channels: []string{}}
	for _, team := range teams {
		if a.api.HasPermissionToTeam(viewerID, team.Id, model.PermissionViewMembers) {
			restrictions.Teams = append(restrictions.Teams, team.Id)
		}
	}

	for page := 0; ; page++ {
		memberships, appErr := a.api.GetChannelMembersForUser("", viewerID, page, membershipsPageSize)
		if appErr != nil {
			return nil, appErr
		}
		for _, m := range memberships {
			restrictions.Channels = append(restrictions.Channels, m.ChannelId)
		}
		if len(memberships) < membershipsPageSize {
			break
		}
	}

	return restrictions, nil
}

// FilterVisibleUsers returns the subset of userIDs allowed by the viewer's
// restrictions, preserving order. Nil restrictions allow every user.
func (a *API) FilterVisibleUsers(viewerID string, restrictions *model.ViewUsersRestrictions, userIDs []string) ([]string, error) {
	if restrictions == nil || len(userIDs) == 0 {
		return userIDs, nil
	}

	visible := map[string]bool{viewerID: true}
	pending := func() []string {
		var ids []string
		for _, id := range userIDs {
			if !visible[id] {
				ids = append(ids, id)
			}
		}
		return ids
	}

	if len(restrictions.Teams) > 0 {
		permittedTeams := make(map[string]bool, len(restrictions.Teams))
		for _, teamID := range restrictions.Teams {
			permittedTeams[teamID] = true
		}
		for _, id := range pending() {
			inTeam, err := a.isActiveMemberOfAnyTeam(id, permittedTeams)
			if err != nil {
				return nil, err
			}
			if inTeam {
				visible[id] = true
			}
		}
	}

	for _, channelID := range restrictions.Channels {
		remaining := pending()
		if len(remaining) == 0 {
			break
		}
		members, appErr := a.api.GetChannelMembersByIds(channelID, remaining)
		if appErr != nil {
			return nil, appErr
		}
		for _, member := range members {
			visible[member.UserId] = true
		}
	}

	result := make([]string, 0, len(userIDs))
	for _, id := range userIDs {
		if visible[id] {
			result = append(result, id)
		}
	}
	return result, nil
}

func (a *API) isActiveMemberOfAnyTeam(userID string, teamIDs map[string]bool) (bool, error) {
	for page := 0; ; page++ {
		members, appErr := a.api.GetTeamMembersForUser(userID, page, membershipsPageSize)
		if appErr != nil {
			return false, appErr
		}
		for _, m := range members {
			if m.DeleteAt == 0 && teamIDs[m.TeamId] {
				return true, nil
			}
		}
		if len(members) < membershipsPageSize {
			return false, nil
		}
	}
}

func (a *API) CleanKVStore() error {
	appErr := a.api.KVDeleteAll()
	if appErr != nil {
		return appErr
	}
	return nil
}

func (a *API) SendEphemeralPost(channelID, mattermostUserID, message string) {
	ephemeralPost := &model.Post{
		ChannelId: channelID,
		UserId:    mattermostUserID,
		Message:   message,
	}
	_ = a.api.SendEphemeralPost(mattermostUserID, ephemeralPost)
}

func (a *API) GetPost(postID string) (*model.Post, error) {
	p, appErr := a.api.GetPost(postID)
	if appErr != nil {
		return nil, appErr
	}
	return p, nil
}

func (a *API) PublishWebsocketEvent(mattermostUserID, event string, payload map[string]any) {
	a.api.PublishWebSocketEvent(event, payload, &model.WebsocketBroadcast{UserId: mattermostUserID})
}

func (a *API) LogAuditRec(rec *model.AuditRecord) {
	a.api.LogAuditRec(rec)
}

func (a *API) GetPreferenceForUser(userID, category, name string) (*model.Preference, error) {
	pref, appErr := a.api.GetPreferenceForUser(userID, category, name)
	if appErr != nil {
		return nil, appErr
	}
	return &pref, nil
}

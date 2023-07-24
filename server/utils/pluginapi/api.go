// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package pluginapi

import (
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

type API struct {
	api plugin.API
}

func New(api plugin.API) *API {
	return &API{
		api: api,
	}
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

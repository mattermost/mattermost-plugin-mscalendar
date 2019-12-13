// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import "github.com/mattermost/mattermost-plugin-msoffice/server/remote"

func (c *client) GetMe() (*remote.User, error) {
	graphUser, err := c.rbuilder.Me().Request().Get(c.ctx)
	if err != nil {
		return nil, err
	}
	user := &remote.User{
		ID:                *graphUser.ID,
		DisplayName:       *graphUser.DisplayName,
		UserPrincipalName: *graphUser.UserPrincipalName,
		Mail: *graphUser.Mail,
	}

	return user, nil
}

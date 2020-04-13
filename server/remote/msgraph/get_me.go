// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

func (c *client) GetMe() (*remote.User, error) {
	graphUser, err := c.rbuilder.Me().Request().Get(c.ctx)
	if err != nil {
		return nil, err
	}

	if graphUser.ID == nil {
		return nil, errors.New("User has no ID")
	}
	if graphUser.DisplayName == nil {
		return nil, errors.New("User has no Display Name")
	}
	if graphUser.UserPrincipalName == nil {
		return nil, errors.New("User has no Principal Name")
	}
	if graphUser.Mail == nil {
		return nil, errors.New("User has no email address. Make sure the Microsoft account is associated to an Outlook product.")
	}

	user := &remote.User{
		ID:                *graphUser.ID,
		DisplayName:       *graphUser.DisplayName,
		UserPrincipalName: *graphUser.UserPrincipalName,
		Mail:              *graphUser.Mail,
	}

	return user, nil
}

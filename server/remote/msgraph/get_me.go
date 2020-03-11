// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

func (c *client) GetMe() (*remote.User, error) {
	graphUser, err := c.rbuilder.Me().Request().Get(c.ctx)
	if err != nil {
		return nil, err
	}

	if graphUser.ID == nil {
		return nil, fmt.Errorf("User has no ID")
	}
	if graphUser.DisplayName == nil {
		return nil, fmt.Errorf("User has no Display Name")
	}
	if graphUser.UserPrincipalName == nil {
		return nil, fmt.Errorf("User has no Principal Name")
	}
	if graphUser.Mail == nil {
		return nil, fmt.Errorf("User has no mail")
	}

	user := &remote.User{
		ID:                *graphUser.ID,
		DisplayName:       *graphUser.DisplayName,
		UserPrincipalName: *graphUser.UserPrincipalName,
		Mail:              *graphUser.Mail,
	}

	return user, nil
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

func (c *client) GetUserEvent(userID, eventID string) (*remote.Event, error) {
	e := &remote.Event{}

	err := c.rbuilder.Users().ID(userID).Events().ID(eventID).Request().JSONRequest(
		c.ctx, http.MethodGet, "", nil, &e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

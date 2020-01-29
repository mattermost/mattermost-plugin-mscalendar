// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

// CreateEvent creates a calendar event
func (c *client) CreateEvent(userID string, in *remote.Event) (*remote.Event, error) {
	var out = remote.Event{}
	err := c.rbuilder.Users().ID(userID).Events().Request().JSONRequest(c.ctx, http.MethodPost, "", &in, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

func (c *client) GetUserCalendars(userID string) ([]*remote.Calendar, error) {
	var v struct {
		Value []*remote.Calendar `json:"value"`
	}
	req := c.rbuilder.Users().ID(userID).Calendars().Request()
	req.Expand("children")
	err := req.JSONRequest(c.ctx, http.MethodGet, "", nil, &v)
	if err != nil {
		return nil, err
	}
	c.LogDebug(fmt.Sprintf("GetUserCalendars: returned %d calendars", len(v.Value)), "userID", userID)
	return v.Value, nil
}

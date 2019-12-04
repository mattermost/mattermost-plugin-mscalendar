// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

// CreateEvent creates a calendar event
func (c *client) CreateEvent(calendarEvent *remote.Event) (*remote.Event, error) {
	var eventOut = remote.Event{}
	req := c.rbuilder.Me().Events().Request()
	err := req.JSONRequest(c.ctx, http.MethodPost, "", &calendarEvent, &eventOut)
	if err != nil {
		return nil, err
	}
	return &eventOut, nil
}

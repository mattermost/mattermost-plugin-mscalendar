// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

// CreateEvent creates a calendar event
func (c *client) CreateEvent(remoteUserID string, in *remote.Event) (*remote.Event, error) {
	var out = remote.Event{}
	err := c.rbuilder.Users().ID(remoteUserID).Events().Request().JSONRequest(c.ctx, http.MethodPost, "", &in, &out)
	if err != nil {
		return nil, errors.Wrap(err, "msgraph CreateEvent")
	}
	return &out, nil
}

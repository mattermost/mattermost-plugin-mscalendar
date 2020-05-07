// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"

	"github.com/larkox/mattermost-plugin-utils/bot/logger"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

// CreateCalendar creates a calendar
func (c *client) CreateCalendar(remoteUserID string, calIn *remote.Calendar) (*remote.Calendar, error) {
	var calOut = remote.Calendar{}
	err := c.rbuilder.Users().ID(remoteUserID).Calendars().Request().JSONRequest(c.ctx, http.MethodPost, "", &calIn, &calOut)
	if err != nil {
		return nil, err
	}
	c.Logger.With(logger.LogContext{
		"v": calOut,
	}).Infof("msgraph: CreateCalendars created the following calendar.")
	return &calOut, nil
}

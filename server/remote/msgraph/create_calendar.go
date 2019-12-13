// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/bot"
)

// CreateCalendar creates a calendar
func (c *client) CreateCalendar(calIn *remote.Calendar) (*remote.Calendar, error) {
	var calOut = remote.Calendar{}
	err := c.rbuilder.Me().Calendars().Request().JSONRequest(c.ctx, http.MethodPost, "", &calIn, &calOut)
	if err != nil {
		return nil, err
	}
	c.Logger.With(bot.LogContext{
		"v": calOut,
	}).Infof("msgraph: CreateCalendars created the following calendar.")
	return &calOut, nil
}

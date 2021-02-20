// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package gcal

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
)

func (c *client) GetCalendars(remoteUserID string) ([]*remote.Calendar, error) {
	if true {
		return nil, errors.New("gcal GetCalendars not implemented")
	}

	var v struct {
		Value []*remote.Calendar `json:"value"`
	}
	req := c.rbuilder.Users().ID(remoteUserID).Calendars().Request()
	req.Expand("children")
	err := req.JSONRequest(c.ctx, http.MethodGet, "", nil, &v)
	if err != nil {
		return nil, errors.Wrap(err, "msgraph GetCalendars")
	}
	c.Logger.With(bot.LogContext{
		"UserID": remoteUserID,
		"v":      v.Value,
	}).Infof("msgraph: GetUserCalendars returned `%d` calendars.", len(v.Value))
	return v.Value, nil
}

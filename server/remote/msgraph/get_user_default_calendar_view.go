// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"
	"net/url"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

func (c *client) GetUserDefaultCalendarView(userID string, start, end time.Time) ([]*remote.Event, error) {
	q := url.Values{}
	q.Add("StartDateTime", start.Format(time.RFC3339))
	q.Add("EndDateTime", end.Format(time.RFC3339))
	params := "?" + q.Encode()

	var v struct {
		Value []*remote.Event `json:"value"`
	}
	err := c.rbuilder.Users().ID(userID).CalendarView().Request().JSONRequest(
		c.ctx, http.MethodGet, params, nil, &v)
	if err != nil {
		return nil, err
	}

	return v.Value, nil
}

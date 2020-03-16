// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/views"
)

func (c *Command) viewCalendar(parameters ...string) (string, error) {
	tz, err := c.MSCalendar.GetTimezone(c.user())
	if err != nil {
		return "Error: No timezone found", err
	}

	events, err := c.MSCalendar.ViewCalendar(c.user(), time.Now(), time.Now().Add(14*24*time.Hour))
	if err != nil {
		return "", err
	}

	return views.RenderCalendarView(events, tz)
}

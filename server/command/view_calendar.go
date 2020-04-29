// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/views"
)

func (c *Command) viewCalendar(parameters ...string) (string, bool, error) {
	tz, err := c.MSCalendar.GetTimezone(c.user())
	if err != nil {
		return "Error: No timezone found", false, err
	}

	events, err := c.MSCalendar.ViewCalendar(c.user(), time.Now().Add(-24*time.Hour), time.Now().Add(14*24*time.Hour))
	if err != nil {
		return "", false, err
	}

	out, err := views.RenderCalendarView(events, tz)
	return out, false, err
}

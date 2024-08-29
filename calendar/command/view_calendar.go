// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine/views"
)

func (c *Command) viewCalendar(_ ...string) (string, bool, error) {
	tz, err := c.Engine.GetTimezone(c.user())
	if err != nil {
		return "Error: No timezone found", false, err
	}

	startOfCurrentDay := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	events, err := c.Engine.ViewCalendar(c.user(), startOfCurrentDay, time.Now().Add(14*24*time.Hour))
	if err != nil {
		return "", false, err
	}

	out, err := views.RenderCalendarView(events, tz)
	return out, false, err
}

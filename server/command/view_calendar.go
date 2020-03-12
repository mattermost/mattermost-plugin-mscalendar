// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"time"
)

func (c *Command) viewCalendar(parameters ...string) (string, error) {
	events, err := c.MSCalendar.ViewCalendar(c.user(), time.Now(), time.Now().Add(14*24*time.Hour))
	if err != nil {
		return "", err
	}

	return c.MSCalendar.RenderCalendarView(c.user(), events)
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
)

func (c *Command) viewCalendar(parameters ...string) (string, error) {
	events, err := c.MSCalendar.ViewCalendar(c.user(), time.Now(), time.Now().Add(14*24*time.Hour))
	if err != nil {
		return "", err
	}

	var timeZone string
	tz, err := c.MSCalendar.GetTimezone(c.user())
	if err == nil {
		timeZone = tz
	}

	if timeZone != "" {
		for _, event := range events {
			event.Start = event.Start.In(timeZone)
			event.End = event.End.In(timeZone)
		}
	}

	resp := ""
	for _, e := range events {
		e.Start = e.Start.In(timeZone)
		e.End = e.End.In(timeZone)
		resp += "  - " + e.ID + utils.JSONBlock(e)
	}

	return resp, nil
}

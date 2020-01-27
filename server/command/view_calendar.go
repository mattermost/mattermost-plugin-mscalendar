// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
)

func (c *Command) viewCalendar(parameters ...string) (string, error) {
	events, err := c.MSCalendar.ViewCalendar(time.Now(), time.Now().Add(14*24*time.Hour))
	if err != nil {
		return "", err
	}

	resp := ""
	for _, e := range events {
		resp += "  - " + e.ID + utils.JSONBlock(e)
	}

	return resp, nil
}

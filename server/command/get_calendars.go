package command

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
)

func (c *Command) showCalendars(parameters ...string) (string, bool, error) {
	resp, err := c.MSCalendar.GetCalendars(c.user())
	if err != nil {
		return "", false, err
	}
	return utils.JSONBlock(resp), false, nil
}

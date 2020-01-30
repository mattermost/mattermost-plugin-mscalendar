package command

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
)

func (c *Command) showCalendars(parameters ...string) (string, error) {

	r, err := c.API.GetUserCalendars("")
	if err != nil {
		return "", err
	}

	return utils.JSONBlock(r), nil
}

package command

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

func (c *Command) createCalendar(parameters ...string) (string, error) {
	if len(parameters) != 1 {
		return "Please provide the name of one calendar to create", nil
	}

	calIn := &remote.Calendar{
		Name: parameters[0],
	}

	_, err := c.MSCalendar.CreateCalendar(c.user(), calIn)
	if err != nil {
		return "", err
	}
	return "", nil
}

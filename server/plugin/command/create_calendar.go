package command

import (
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

func (c *Command) createCalendar(parameters ...string) (string, error) {
	if len(parameters) != 1 {
		return "Please provide the name of one calendar to create", nil
	}

	calIn := &remote.Calendar{
		Name: parameters[0],
	}

	_, err := c.API.CreateCalendar(calIn)
	if err != nil {
		return "", err
	}
	return "", nil
}

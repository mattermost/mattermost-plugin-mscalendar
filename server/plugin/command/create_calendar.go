package command

import (
	"fmt"
)

func (c *Command) createCalendar(parameters ...string) (string, error) {
	if len(parameters) != 1 {
		return "Please provide the name of one calendar to create", nil
	}

	calendar, err := c.API.CreateCalendar(parameters[0])
	if err != nil {
		return "", err
	}
	fmt.Printf("calendar = %+v\n", calendar)
	return "", nil
}

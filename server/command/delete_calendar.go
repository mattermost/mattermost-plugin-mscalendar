package command

func (c *Command) deleteCalendar(parameters ...string) (string, bool, error) {
	if len(parameters) != 1 {
		return "Please provide the ID of only one calendar ", false, nil
	}

	err := c.MSCalendar.DeleteCalendar(c.user(), parameters[0])
	if err != nil {
		return "", false, err
	}
	return "", false, nil
}

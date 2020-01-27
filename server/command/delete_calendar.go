package command

func (c *Command) deleteCalendar(parameters ...string) (string, error) {
	if len(parameters) != 1 {
		return "Please provide the ID of only one calendar ", nil
	}

	err := c.MSCalendar.DeleteCalendar(parameters[0])
	if err != nil {
		return "", err
	}
	return "", nil
}

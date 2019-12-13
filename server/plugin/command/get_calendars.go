package command

func (c *Command) showCalendars(parameters ...string) (string, error) {

	_, err := c.API.GetUserCalendars("")
	if err != nil {
		return "", err
	}

	return "", nil
}

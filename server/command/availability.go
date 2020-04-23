// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

func (c *Command) availability(parameters ...string) (string, error) {
	switch {
	case len(parameters) == 0:
		resString, err := c.MSCalendar.Sync(c.Args.UserId)
		if err != nil {
			return "", err
		}

		return resString, nil
	case len(parameters) == 1 && parameters[0] == "all":
		resString, err := c.MSCalendar.SyncAll()
		if err != nil {
			return "", err
		}

		return resString, nil
	}

	return "bad syntax", nil
}

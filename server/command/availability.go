// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

func (c *Command) availability(parameters ...string) (string, bool, error) {
	authorized, err := c.MSCalendar.IsAuthorizedAdmin(c.Args.UserId)
	if err != nil {
		return "", false, err
	}
	if !authorized {
		return "Not authorized", false, nil
	}

	switch {
	case len(parameters) == 0:
		resString, err := c.MSCalendar.Sync(c.Args.UserId)
		if err != nil {
			return "", false, err
		}

		return resString, false, nil
	case len(parameters) == 1 && parameters[0] == "all":
		resString, err := c.MSCalendar.SyncAll()
		if err != nil {
			return "", false, err
		}

		return resString, false, nil
	}

	return "bad syntax", false, nil
}

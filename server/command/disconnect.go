// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

func (c *Command) disconnect(parameters ...string) (string, error) {
	err := c.MSCalendar.DisconnectUser(c.Args.UserId)
	if err != nil {
		return "", err
	}

	return "Successfully disconnected your account", nil
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import "github.com/pkg/errors"

func (c *Command) disconnect(parameters ...string) (string, error) {
	err := c.MSCalendar.DisconnectUser(c.Args.UserId)
	if err != nil {
		return "", err
	}

	return "Successfully disconnected your account", nil
}

func (c *Command) disconnectBot(parameters ...string) (string, error) {
	isAdmin, err := c.MSCalendar.IsAuthorizedAdmin(c.Args.UserId)
	if err != nil || !isAdmin {
		return "", errors.New("non-admin user attempting to disconnect bot account")
	}

	err = c.MSCalendar.DisconnectUser(c.Config.BotUserID)
	if err != nil {
		return "", err
	}

	return "Successfully disconnected bot user", nil
}

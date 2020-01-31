// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

func (c *Command) disconnect(parameters ...string) (string, error) {
	err := c.API.DisconnectUser(c.Args.UserId)
	if err != nil {
		return "", err
	}

	return "Successfully disconnected your account", nil
}

func (c *Command) disconnectBot(parameters ...string) (string, error) {
	err := c.API.DisconnectBot()
	if err != nil {
		return "", err
	}

	return "Successfully disconnected bot user", nil
}

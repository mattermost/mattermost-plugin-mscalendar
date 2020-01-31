// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
)

func (c *Command) connect(parameters ...string) (string, error) {
	_, err := c.API.GetRemoteUser(c.Args.UserId)
	if err == nil {
		return "Your account is already connected. Please run `/mscalendar disconnect`", nil
	}

	out := fmt.Sprintf("[Click here to link your %s account.](%s/oauth2/connect)",
		config.ApplicationName,
		c.Config.PluginURL)
	return out, nil
}

func (c *Command) connectBot(parameters ...string) (string, error) {
	_, err := c.API.GetRemoteUser(c.Config.BotUserID)
	if err == nil {
		return "Bot user already connected. Please run `/mscalendar disconnect_bot`", nil
	}

	out := fmt.Sprintf("[Click here to link the bot's %s account.](%s/oauth2/connect_bot)",
		config.ApplicationName,
		c.Config.PluginURL)
	return out, nil
}

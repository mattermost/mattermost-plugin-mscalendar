// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/pkg/errors"
)

func (c *Command) connect(parameters ...string) (string, error) {
	u, err := c.API.GetRemoteUser(c.Args.UserId) // needs fix
	if err == nil {
		return fmt.Sprintf("Your account is already connected to %s. Please run `/mscalendar disconnect`", u.Remote.Mail), nil
	}

	out := fmt.Sprintf("[Click here to link your %s account.](%s/oauth2/connect)",
		config.ApplicationName,
		c.Config.PluginURL)
	return out, nil
}

func (c *Command) connectBot(parameters ...string) (string, error) {
	isAdmin, err := c.API.IsAuthorizedAdmin(c.Args.UserId) // needs fix
	if err != nil || !isAdmin {
		return "", errors.New("Non-admin user attempting to connect bot account")
	}

	_, err = c.API.GetRemoteUser(c.Config.BotUserID) // needs fix
	if err == nil {
		return "Bot user already connected. Please run `/mscalendar disconnect_bot`", nil
	}

	out := fmt.Sprintf("[Click here to link the bot's %s account.](%s/oauth2/connect?bot=true)",
		config.ApplicationName,
		c.Config.PluginURL)
	return out, nil
}

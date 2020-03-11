// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/pkg/errors"
)

func (c *Command) connect(parameters ...string) (string, error) {
	ru, err := c.MSCalendar.GetRemoteUser(c.Args.UserId)
	if err == nil {
		return fmt.Sprintf("Your Mattermost account is already connected to %s account `%s`. To connect to a different account, first run `/%s disconnect`.", config.ApplicationName, ru.Mail, config.CommandTrigger), nil
	}

	out := fmt.Sprintf(`Welcome to the Microsoft Calendar Bot.
	[Click here to link your account.](%s/oauth2/connect)`, c.Config.PluginURL)

	err = c.MSCalendar.Welcome(c.Args.UserId)
	if err != nil {
		out = "There has been a problem while trying to connect. err=" + err.Error()
	}

	return out, nil
}

func (c *Command) connectBot(parameters ...string) (string, error) {
	isAdmin, err := c.MSCalendar.IsAuthorizedAdmin(c.Args.UserId)
	if err != nil || !isAdmin {
		return "", errors.New("non-admin user attempting to connect bot account")
	}

	ru, err := c.MSCalendar.GetRemoteUser(c.Config.BotUserID)
	if err == nil {
		return fmt.Sprintf("The bot account is already connected to %s account `%s`. To connect to a different account, first run `/%s disconnect_bot`.", config.ApplicationName, ru.Mail, config.CommandTrigger), nil
	}

	out := fmt.Sprintf("[Click here to link the bot's %s account.](%s/oauth2/connect_bot)",
		config.ApplicationName,
		c.Config.PluginURL)
	return out, nil
}

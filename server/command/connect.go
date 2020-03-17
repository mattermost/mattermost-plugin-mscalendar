// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
	"github.com/pkg/errors"
)

const (
	ConnectBotAlreadyConnectedTemplate = "The bot account is already connected to %s account `%s`. To connect to a different account, first run `/%s disconnect_bot`."
	ConnectBotSuccessTemplate          = "[Click here to link the bot's %s account.](%s/oauth2/connect_bot)"
	ConnectAlreadyConnectedTemplate    = "Your Mattermost account is already connected to %s account `%s`. To connect to a different account, first run `/%s disconnect`."
	ConnectErrorMessage                = "There has been a problem while trying to connect. err="
)

func (c *Command) connect(parameters ...string) (string, error) {
	ru, err := c.MSCalendar.GetRemoteUser(c.Args.UserId)
	if err == nil {
		return fmt.Sprintf(ConnectAlreadyConnectedTemplate, config.ApplicationName, ru.Mail, config.CommandTrigger), nil
	}

	out := fmt.Sprintf(mscalendar.ConnectSuccessTemplate, c.Config.PluginURL)

	err = c.MSCalendar.Welcome(c.Args.UserId)
	if err != nil {
		out = ConnectErrorMessage + err.Error()
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
		return fmt.Sprintf(ConnectBotAlreadyConnectedTemplate, config.ApplicationName, ru.Mail, config.CommandTrigger), nil
	}

	out := fmt.Sprintf(ConnectBotSuccessTemplate,
		config.ApplicationName,
		c.Config.PluginURL)
	return out, nil
}

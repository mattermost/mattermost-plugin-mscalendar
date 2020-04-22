// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
)

func (c *Command) connect(parameters ...string) (string, error) {
	ru, err := c.MSCalendar.GetRemoteUser(c.Args.UserId)
	if err == nil {
		return fmt.Sprintf("Your Mattermost account is already connected to %s account `%s`. To connect to a different account, first run `/%s disconnect`.", config.ApplicationName, ru.Mail, config.CommandTrigger), nil
	}

	out := fmt.Sprintf("[Click here to link your %s account.](%s/oauth2/connect)",
		config.ApplicationName,
		c.Config.PluginURL)
	return out, nil
}

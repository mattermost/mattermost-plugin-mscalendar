// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
)

func (c *Command) help(parameters ...string) (string, bool, error) {
	resp := fmt.Sprintf("Mattermost Microsoft Calendar plugin version: %s, "+
		"[%s](https://github.com/mattermost/%s/commit/%s), built %s\n",
		c.Config.PluginVersion,
		c.Config.BuildHashShort,
		config.Repository,
		c.Config.BuildHash,
		c.Config.BuildDate)
	resp += "\n"
	resp += getCommandText("")
	resp += getCommandText("help")
	resp += getCommandText("connect")
	resp += getCommandText("disconnect")
	resp += getCommandText("settings - Edit your user personal settings.")
	resp += getCommandText("summary - View your events for today, or edit the settings for your daily summary.")
	resp += getCommandText("viewcal - View your events for the upcoming week.")
	resp += getCommandText("subscribe - Enable notifications for event invitations and updates.")
	resp += getCommandText("unsubscribe - Disable notifications for event invitations and updates.")
	return resp, false, nil
}

func getCommandText(s string) string {
	return fmt.Sprintf("/%s %s\n", config.CommandTrigger, s)
}

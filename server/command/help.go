// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
)

func (c *Command) help(parameters ...string) (string, error) {
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
	resp += getCommandText("info")
	resp += getCommandText("connect")
	resp += getCommandText("viewcal")
	resp += getCommandText("showcals")
	resp += getCommandText("subscribe")
	resp += getCommandText("unsubscribe")
	resp += getCommandText("createcal <name>")
	resp += getCommandText("deletecal <id>")
	resp += getCommandText("createevent")
	resp += getCommandText("findmeetings (Optional: <attendees>)")
	resp += "  * <attendees> - space delimited <type>:<email> combinations \n"
	resp += "  * <type> options - required, optional \n"
	return resp, nil
}

func getCommandText(s string) string {
	return fmt.Sprintf("/%s %s\n", config.CommandTrigger, s)
}

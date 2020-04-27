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
	resp += fmt.Sprintf("* /%s\n", config.CommandTrigger)
	resp += fmt.Sprintf("* /%s help\n", config.CommandTrigger)
	resp += fmt.Sprintf("* /%s info\n", config.CommandTrigger)
	resp += fmt.Sprintf("* /%s connect\n", config.CommandTrigger)
	resp += fmt.Sprintf("* /%s viewcal\n", config.CommandTrigger)
	resp += fmt.Sprintf("* /%s showcals\n", config.CommandTrigger)
	resp += fmt.Sprintf("* /%s subscribe\n", config.CommandTrigger)
	resp += fmt.Sprintf("* /%s unsubscribe\n", config.CommandTrigger)
	resp += fmt.Sprintf("* /%s createcal <name>\n", config.CommandTrigger)
	resp += fmt.Sprintf("* /%s deletecal <id>\n", config.CommandTrigger)
	resp += fmt.Sprintf("* /%s createevent\n", config.CommandTrigger)
	resp += fmt.Sprintf("* /%s findmeetings (Optional: <attendees>)\n", config.CommandTrigger)
	resp += "  * <attendees> - space delimited <type>:<email> combinations \n"
	resp += "  * <type> options - required, optional \n"
	return resp, nil
}

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
	resp += "* /mscalendar\n"
	resp += "* /mscalendar help\n"
	resp += "* /mscalendar info\n"
	resp += "* /mscalendar connect\n"
	resp += "* /mscalendar viewcal\n"
	resp += "* /mscalendar showcals\n"
	resp += "* /mscalendar subscribe\n"
	resp += "* /mscalendar createcal <name>\n"
	resp += "* /mscalendar deletecal <id>\n"
	resp += "* /mscalendar createevent\n"
	resp += "* /mscalendar findmeetings (Optional: <attendees>)\n"
	resp += "  * <attendees> - space delimited <type>:<email> combinations \n"
	resp += "  * <type> options - required, optional \n"
	return resp, nil
}

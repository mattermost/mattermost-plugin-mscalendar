// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
)

func (c *Command) help(parameters ...string) (string, error) {
	resp := fmt.Sprintf("Mattermost Microsoft Office plugin version: %s, "+
		"[%s](https://github.com/mattermost/%s/commit/%s), built %s\n",
		c.Config.PluginVersion,
		c.Config.BuildHashShort,
		config.Repository,
		c.Config.BuildHash,
		c.Config.BuildDate)
	resp += "\n"
	resp += "TODO help\n"
	resp += "/msoffice\n"
	resp += "/msoffice connect\n"
	resp += "/msoffice help\n"
	resp += "/msoffice info\n"
	resp += "/msoffice viewcal\n"
	resp += "/msoffice subscribe\n"
	resp += "/msoffice createcal <name>\n"
	resp += "/msoffice deletecal <id>\n"
	resp += "/msoffice createevent <Subject> <Start> <End> <Reminder (minutes)>\n"
	return resp, nil
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
)

func (c *Command) help(parameters ...string) (string, bool, error) {
	resp := ""
	for _, cmd := range cmds {
		desc := cmd.Trigger
		if cmd.HelpText != "" {
			desc += " - " + cmd.HelpText
		}
		resp += getCommandText(desc)
	}
	return resp, false, nil
}

func getCommandText(s string) string {
	return fmt.Sprintf("/%s %s\n", config.CommandTrigger, s)
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
)

func (c *Command) info(parameters ...string) (string, error) {
	resp := fmt.Sprintf("Mattermost Microsoft Office plugin version: %s, "+
		"[%s](https://github.com/mattermost/%s/commit/%s), built %s\n",
		c.Config.PluginVersion,
		c.Config.BuildHashShort,
		config.Repository,
		c.Config.BuildHash,
		c.Config.BuildDate)
	return resp, nil
}

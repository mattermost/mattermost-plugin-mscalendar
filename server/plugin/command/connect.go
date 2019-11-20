// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
)

func (c *Command) connect(parameters ...string) (string, error) {
	out := fmt.Sprintf("[Click here to link your %s account.](%s/oauth2/connect)",
		config.ApplicationName,
		c.Config.PluginURL)
	return out, nil
}

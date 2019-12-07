// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package plugin

import (
	"strings"

	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/bot"
)

// isAuthorized returns true if the user is authorized to use the workflow plugin's admin-level APIs/commands.
func (p *Plugin) IsAuthorizedAdmin(mattermostUserID string) (bool, error) {
	user, err := p.API.GetUser(mattermostUserID)
	if err != nil {
		return false, err
	}
	if strings.Contains(user.Roles, "system_admin") {
		return true, nil
	}
	conf := p.getConfig()
	bot := bot.NewBot(p.API, conf.BotUserID).WithConfig(conf.BotConfig)
	return bot.IsUserAdmin(mattermostUserID), nil
}

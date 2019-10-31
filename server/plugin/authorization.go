// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package plugin

import (
	"strings"
)

// isAuthorized returns true if the user is authorized to use the workflow plugin's admin-level APIs/commands.
func (p *Plugin) IsAuthorizedAdmin(mattermostID string) (bool, error) {
	user, err := p.API.GetUser(mattermostID)
	if err != nil {
		return false, err
	}
	if strings.Contains(user.Roles, "system_admin") { // || p.userAuthorized(mattermostID) {
		return true, nil
	}
	return false, nil
}

// func (p *Plugin) userAuthorized(userName string) bool {
// 	list := strings.Split(p.getConfig().AllowedUserIDs, ",")
// 	for _, u := range list {
// 		if userName == strings.TrimSpace(u) {
// 			return true
// 		}
// 	}
// 	return false
// }

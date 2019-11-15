// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
)

func (r *Handler) help(parameters ...string) (string, error) {
	resp := fmt.Sprintf("Mattermost Microsoft Office plugin version: %s, "+
		"[%s](https://github.com/mattermost/%s/commit/%s), built %s\n",
		r.Config.PluginVersion,
		r.Config.BuildHashShort,
		config.Repository,
		r.Config.BuildHash,
		r.Config.BuildDate)
	resp += "\n"
	resp += "TODO help\n"
	resp += "/msoffice connect\n"
	resp += "/msoffice viewcal\n"
	resp += "/msoffice createmeeting <Subject> <Body> <CalendarId>\n"
	// resp += "/msoffice createmeeting <Subject> <Body> <CalendarId> <Start> <End>\n"
	resp += "/msoffice createcalendar <calendar_name>\n"
	resp += "/msoffice calgetevents\n"
	return resp, nil
}

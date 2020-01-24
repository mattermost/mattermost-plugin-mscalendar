// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

func (c *client) GetUserMailboxSettings(remoteUserID string) (*remote.MailboxSettings, error) {
	var u string
	if remoteUserID == "me" {
		u = "/me/mailboxSettings"
	} else {
		u = "/users/" + remoteUserID + "/mailboxSettings"
	}

	out := &remote.MailboxSettings{}

	_, err := c.CallJSON(http.MethodGet, u, nil, out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

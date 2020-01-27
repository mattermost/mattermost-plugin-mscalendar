// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

func (c *client) GetUserMailboxSettings(remoteUserID string) (*remote.MailboxSettings, error) {
	u := "/users/" + remoteUserID + "/mailboxSettings"
	out := &remote.MailboxSettings{}

	_, err := c.CallJSON(http.MethodGet, u, nil, out)
	return out, err
}

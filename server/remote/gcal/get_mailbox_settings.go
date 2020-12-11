// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package gcal

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

func (c *client) GetMailboxSettings(remoteUserID string) (*remote.MailboxSettings, error) {
	// GCAL TODO
	out := &remote.MailboxSettings{
		TimeZone: "Eastern Standard Time",
	}
	return out, nil
}

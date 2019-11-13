// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package user

import (
	"fmt"

	graph "github.com/jkrecek/msgraph-go"
	"golang.org/x/oauth2"
)

type User struct {
	graph.Me
	PluginVersion string
	OAuth2Token   *oauth2.Token `json:",omitempty"`
	Settings      *Settings
}

type Settings struct {
	Notifications bool `json:"notifications"`
}

func (settings Settings) String() string {
	notifications := "off"
	if settings.Notifications {
		notifications = "on"
	}
	return fmt.Sprintf("\tNotifications: %s", notifications)
}

// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"fmt"

	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

type User struct {
	PluginVersion    string
	MattermostUserID string
	Remote           *remote.User
	OAuth2Token      *oauth2.Token `json:",omitempty"`
	Settings         Settings
}

type Settings struct {
	EventSubscriptionID string
}

func (settings Settings) String() string {
	sub := "no subscription"
	if settings.EventSubscriptionID != "" {
		sub = "subscription ID: " + settings.EventSubscriptionID
	}
	return fmt.Sprintf(" - %s", sub)
}

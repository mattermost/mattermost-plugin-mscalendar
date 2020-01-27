// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/plugin_api"
)

type MSCalendar interface {
	Availability
	Calendar
	Client
	Event
	OAuth2
	Subscriptions
}

// Dependencies contains all API dependencies
type Dependencies struct {
	EventStore        store.EventStore
	IsAuthorizedAdmin func(userId string) (bool, error)
	Logger            bot.Logger
	OAuth2StateStore  store.OAuth2StateStore
	PluginAPI         plugin_api.PluginAPI
	Poster            bot.Poster
	Remote            remote.Remote
	SubscriptionStore store.SubscriptionStore
	UserStore         store.UserStore
}

type Config struct {
	*Dependencies
	*config.Config
}

type mscalendar struct {
	Config
	mattermostUserID string
	user             *store.User
}

func New(apiConfig Config, mattermostUserID string) MSCalendar {
	return &mscalendar{
		Config:           apiConfig,
		mattermostUserID: mattermostUserID,
	}
}

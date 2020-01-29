// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/oauth2connect"
)

type MSCalendar interface {
	Availability
	Calendar
	EventResponder
	oauth2connect.App
	Subscriptions
	Users
}

type PluginAPI interface {
	GetMattermostUser(mattermostUserID string) (*model.User, error)
	GetMattermostUserByUsername(mattermostUsername string) (*model.User, error)
	GetMattermostUserStatus(userID string) (*model.Status, error)
	GetMattermostUserStatusesByIds(userIDs []string) ([]*model.Status, error)
	IsSysAdmin(mattermostUserID string) (bool, error)
	UpdateMattermostUserStatus(userID, status string) (*model.Status, error)
}

// Dependencies contains all API dependencies
type Dependencies struct {
	EventStore        store.EventStore
	Logger            bot.Logger
	OAuth2StateStore  store.OAuth2StateStore
	PluginAPI         PluginAPI
	Poster            bot.Poster
	Remote            remote.Remote
	SubscriptionStore store.SubscriptionStore
	UserStore         store.UserStore
}

type Env struct {
	*Dependencies
	*config.Config
}

type mscalendar struct {
	Env

	actingUser *User
	client     remote.Client
}

func New(env Env, actingMattermostUserID string) MSCalendar {
	if actingMattermostUserID == "" {
		actingMattermostUserID = env.Config.BotUserID
	}
	return &mscalendar{
		Env:        env,
		actingUser: NewUser(actingMattermostUserID),
	}
}

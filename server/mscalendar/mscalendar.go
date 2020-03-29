// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
)

type MSCalendar interface {
	Availability
	Calendar
	EventResponder
	Subscriptions
	Users
	DailySummary
}

// Dependencies contains all API dependencies
type Dependencies struct {
	Logger            bot.Logger
	PluginAPI         PluginAPI
	Poster            bot.Poster
	Remote            remote.Remote
	Store             store.Store
	IsAuthorizedAdmin func(string) (bool, error)
}

type PluginAPI interface {
	GetMattermostUser(mattermostUserID string) (*model.User, error)
	GetMattermostUserByUsername(mattermostUsername string) (*model.User, error)
	GetMattermostUserStatusesByIds(mattermostUserIDs []string) ([]*model.Status, error)
	IsSysAdmin(mattermostUserID string) (bool, error)
	UpdateMattermostUserStatus(mattermostUserID, status string) (*model.Status, error)
}

type Env struct {
	*config.Config
	*Dependencies
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

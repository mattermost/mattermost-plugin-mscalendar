// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/larkox/mattermost-plugin-utils/bot/logger"
	"github.com/larkox/mattermost-plugin-utils/bot/poster"
	"github.com/larkox/mattermost-plugin-utils/panel"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

type MSCalendar interface {
	Availability
	Calendar
	EventResponder
	Subscriptions
	Users
	Welcomer
	Settings
	DailySummary
}

// Dependencies contains all API dependencies
type Dependencies struct {
	Logger            logger.Logger
	PluginAPI         PluginAPI
	Poster            poster.Poster
	Remote            remote.Remote
	Store             store.Store
	SettingsPanel     panel.Panel
	IsAuthorizedAdmin func(string) (bool, error)
	Welcomer          Welcomer
}

type PluginAPI interface {
	GetMattermostUser(mattermostUserID string) (*model.User, error)
	GetMattermostUserByUsername(mattermostUsername string) (*model.User, error)
	GetMattermostUserStatusesByIds(mattermostUserIDs []string) ([]*model.Status, error)
	IsSysAdmin(mattermostUserID string) (bool, error)
	UpdateMattermostUserStatus(mattermostUserID, status string) (*model.Status, error)
	GetPost(postID string) (*model.Post, error)
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
	return &mscalendar{
		Env:        env,
		actingUser: NewUser(actingMattermostUserID),
	}
}

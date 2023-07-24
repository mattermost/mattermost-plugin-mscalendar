// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/tracker"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/settingspanel"
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
	Logger            bot.Logger
	PluginAPI         PluginAPI
	Poster            bot.Poster
	Remote            remote.Remote
	Store             store.Store
	SettingsPanel     settingspanel.Panel
	IsAuthorizedAdmin func(string) (bool, error)
	Welcomer          Welcomer
	Tracker           tracker.Tracker
}

type PluginAPI interface {
	GetMattermostUser(mattermostUserID string) (*model.User, error)
	GetMattermostUserByUsername(mattermostUsername string) (*model.User, error)
	GetMattermostUserStatus(mattermostUserID string) (*model.Status, error)
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

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package bot

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type Bot interface {
	Poster
	Logger
	Admin

	WithConfig(BotConfig) Bot
	UserID() string
}

type bot struct {
	BotConfig
	pluginAPI        plugin.API
	mattermostUserID string
	displayName      string
	logContext       LogContext
}

func Ensure(api plugin.API, helpers plugin.Helpers, stored *model.Bot, iconPath string) (Bot, string, error) {
	botUserID, err := helpers.EnsureBot(stored, plugin.ProfileImagePath(iconPath))
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to ensure bot account")
	}

	bot := &bot{
		pluginAPI:        api,
		mattermostUserID: botUserID,
		displayName:      stored.DisplayName,
	}
	return bot, botUserID, nil
}

func (bot *bot) WithConfig(conf BotConfig) Bot {
	newbot := *bot
	newbot.BotConfig = conf
	return &newbot
}

func (bot *bot) UserID() string {
	return bot.mattermostUserID
}

func (bot *bot) String() string {
	return bot.displayName
}

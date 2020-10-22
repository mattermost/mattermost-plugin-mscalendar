// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package bot

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/flow"
)

type Bot interface {
	Poster
	Logger
	Admin
	FlowController

	Ensure(stored *model.Bot, iconPath string) error
	WithConfig(Config) Bot
	MattermostUserID() string
	RegisterFlow(flow.Flow, flow.Store)
}

type bot struct {
	Config
	pluginAPI        plugin.API
	pluginHelpers    plugin.Helpers
	mattermostUserID string
	displayName      string
	logContext       LogContext
	pluginURL        string

	flow      flow.Flow
	flowStore flow.Store
}

func New(api plugin.API, helpers plugin.Helpers, pluginURL string) Bot {
	return &bot{
		pluginAPI:     api,
		pluginHelpers: helpers,
		pluginURL:     pluginURL,
	}
}

func (bot *bot) RegisterFlow(flow flow.Flow, flowStore flow.Store) {
	bot.flow = flow
	bot.flowStore = flowStore
}

func (bot *bot) Ensure(stored *model.Bot, iconPath string) error {
	if bot.mattermostUserID != "" {
		// Already done
		return nil
	}

	botUserID, err := bot.pluginHelpers.EnsureBot(stored, plugin.ProfileImagePath(iconPath))
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}
	bot.mattermostUserID = botUserID
	bot.displayName = stored.DisplayName
	return nil
}

func (bot *bot) WithConfig(conf Config) Bot {
	newbot := *bot
	newbot.Config = conf
	return &newbot
}

func (bot *bot) MattermostUserID() string {
	return bot.mattermostUserID
}

func (bot *bot) String() string {
	return bot.displayName
}

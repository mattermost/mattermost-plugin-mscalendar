// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package bot

import (
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type Bot interface {
	Poster
	Logger
	Admin

	WithConfig(BotConfig) Bot
}

type BotConfig struct {
	// AdminUserIDs contains a comma-separated list of user IDs that are allowed
	// to administer plugin functions, even if not Mattermost sysadmins.
	AdminUserIDs string

	// AdminLogLevel is "debug", "info", "warn", or "error".
	AdminLogLevel string

	// AdminLogVerbose: set to include full context with admin log messages.
	AdminLogVerbose bool
}

func (c BotConfig) ToStorableConfig(configMap map[string]interface{}) map[string]interface{} {
	if configMap == nil {
		configMap = map[string]interface{}{}
	}
	configMap["AdminUserIDs"] = c.AdminUserIDs
	configMap["AdminLogLevel"] = c.AdminLogLevel
	configMap["AdminLogVerbose"] = c.AdminLogVerbose
	return configMap
}

type bot struct {
	BotConfig
	pluginAPI        plugin.API
	mattermostUserID string
	logContext       LogContext
}

// NewBot creates a new bot poster.
func NewBot(api plugin.API, botUserID string) Bot {
	return &bot{
		pluginAPI:        api,
		mattermostUserID: botUserID,
	}
}

func (bot *bot) WithConfig(conf BotConfig) Bot {
	newbot := *bot
	newbot.BotConfig = conf
	return &newbot
}

package mscalendar

import (
	"github.com/larkox/mattermost-plugin-utils/bot"
	"github.com/larkox/mattermost-plugin-utils/bot/logger"
	"github.com/larkox/mattermost-plugin-utils/bot/poster"
	"github.com/larkox/mattermost-plugin-utils/flow"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type Bot interface {
	poster.Poster
	logger.Logger
	bot.Admin
	bot.Bot
	flow.FlowController
	Welcomer
}

type msCalendarBot struct {
	bot.Bot
	poster.Poster
	bot.Admin
	logger.Logger
	flow.FlowController
	Welcomer
}

func NewBot(helpers plugin.Helpers, api plugin.API, adminUserIDs string, logConfig logger.Config, pluginURL string, env Env) Bot {
	newBot := &msCalendarBot{}
	newBot.Bot = bot.New(helpers)
	newBot.Poster = poster.NewPoster(api, newBot.MattermostUserID())
	newBot.Admin = bot.NewAdmin(adminUserIDs, newBot)
	newBot.Logger = logger.NewLogger(logConfig, newBot, newBot, api)
	newBot.FlowController = flow.NewFlowController(newBot, newBot, pluginURL)
	newBot.Welcomer = NewWelcomer(newBot, env, pluginURL)

	return newBot
}

func (bot *msCalendarBot) Ensure(stored *model.Bot, iconPath string) error {
	err := bot.Bot.Ensure(stored, iconPath)
	if err != nil {
		return err
	}

	bot.Poster.UpdatePosterID(bot.MattermostUserID())

	return nil
}

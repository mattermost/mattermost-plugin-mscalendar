package mscalendar

import (
	"fmt"

	"github.com/larkox/mattermost-plugin-utils/bot/logger"
	"github.com/larkox/mattermost-plugin-utils/bot/poster"
	"github.com/larkox/mattermost-plugin-utils/flow"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-server/v5/model"
)

type Welcomer interface {
	Welcome(userID string) error
	AfterSuccessfullyConnect(userID, userLogin string) error
	AfterDisconnect(userID string) error
	WelcomeFlowEnd(userID string)
}

type welcomer struct {
	poster.Poster
	logger.Logger
	flow.FlowController
	Env
	pluginURL string
}

const (
	WelcomeMessage = `Welcome to the Microsoft Calendar plugin.
	[Click here to link your account.](%s/oauth2/connect)`
)

func (m *mscalendar) Welcome(userID string) error {
	return m.Welcomer.Welcome(userID)
}

func (m *mscalendar) AfterSuccessfullyConnect(userID, userLogin string) error {
	return m.Welcomer.AfterSuccessfullyConnect(userID, userLogin)
}

func (m *mscalendar) AfterDisconnect(userID string) error {
	return m.Welcomer.AfterDisconnect(userID)
}

func (m *mscalendar) WelcomeFlowEnd(userID string) {
	m.Welcomer.WelcomeFlowEnd(userID)
}

func NewWelcomer(bot Bot, env Env, pluginURL string) Welcomer {
	return &welcomer{
		Poster:         bot,
		Logger:         bot,
		FlowController: bot,
		Env:            env,
		pluginURL:      pluginURL,
	}
}

func (bot *welcomer) Welcome(userID string) error {
	bot.cleanWelcomePost(userID)

	postID, err := bot.DMWithAttachments(userID, bot.newConnectAttachment())
	if err != nil {
		return err
	}

	bot.Store.StoreUserWelcomePost(userID, postID)

	return nil
}

func (bot *welcomer) AfterSuccessfullyConnect(userID, userLogin string) error {
	postID, err := bot.Store.DeleteUserWelcomePost(userID)
	if err != nil {
		bot.Errorf("error deleting user welcom post id, err=" + err.Error())
	}
	if postID != "" {
		post := &model.Post{
			Id: postID,
		}
		model.ParseSlackAttachment(post, []*model.SlackAttachment{bot.newConnectedAttachment(userLogin)})
		bot.UpdatePost(post)
	}

	return bot.Start(userID)
}

func (bot *welcomer) AfterDisconnect(userID string) error {
	errCancel := bot.Cancel(userID)
	errClean := bot.cleanWelcomePost(userID)
	if errCancel != nil {
		return errCancel
	}

	if errClean != nil {
		return errClean
	}
	return nil
}

func (bot *welcomer) WelcomeFlowEnd(userID string) {
	bot.notifySettings(userID)
}

func (bot *welcomer) newConnectAttachment() *model.SlackAttachment {
	sa := model.SlackAttachment{
		Title: "Connect",
		Text:  fmt.Sprintf(WelcomeMessage, bot.pluginURL),
	}

	return &sa
}

func (bot *welcomer) newConnectedAttachment(userLogin string) *model.SlackAttachment {
	return &model.SlackAttachment{
		Title: "Connect",
		Text:  ":tada: Congratulations! Your microsoft account (*" + userLogin + "*) has been connected to Mattermost.",
	}
}

func (bot *welcomer) notifySettings(userID string) error {
	_, err := bot.DM(userID, "Feel free to change these settings anytime by typing `/%s settings`", config.CommandTrigger)
	if err != nil {
		return err
	}
	return nil
}

func (bot *welcomer) cleanWelcomePost(mattermostUserID string) error {
	postID, err := bot.Store.DeleteUserWelcomePost(mattermostUserID)
	if err != nil {
		return err
	}

	if postID != "" {
		err = bot.DeletePost(postID)
		if err != nil {
			bot.Errorf(err.Error())
		}
	}
	return nil
}

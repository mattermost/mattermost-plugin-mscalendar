package mscalendar

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/flow"
	"github.com/mattermost/mattermost-server/v5/model"
)

type Welcomer interface {
	Welcome(userID string) error
	AfterSuccessfullyConnect(userID, userLogin string) error
	AfterDisconnect(userID string) error
	WelcomeFlowEnd(userID string)
}

type Bot interface {
	bot.Bot
	Welcomer
	flow.FlowStore
}

type mscBot struct {
	bot.Bot
	Env
	pluginURL string
}

const (
	WelcomeMessage = `### Welcome to the Microsoft Calendar plugin!
Here is some info to prove we got you logged in
- Name: %s
`
	ConnectSuccessTemplate = `Welcome to the Microsoft Calendar.
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

func GetMSCalendarBot(bot bot.Bot, env Env, pluginURL string) Bot {
	return &mscBot{
		Bot:       bot,
		Env:       env,
		pluginURL: pluginURL,
	}
}

func (bot *mscBot) Welcome(userID string) error {
	bot.cleanPostIDs(userID)

	postID, err := bot.DMWithAttachments(userID, bot.getConnectAttachment())
	if err != nil {
		return err
	}

	bot.Store.StoreUserWelcomePost(userID, postID)

	return nil
}

func (bot *mscBot) AfterSuccessfullyConnect(userID, userLogin string) error {
	postID, err := bot.Store.DeleteUserWelcomePost(userID)
	if err != nil {
		bot.Errorf("error deleting user welcom post id, err=" + err.Error())
	}
	if postID != "" {
		post := &model.Post{
			Id: postID,
		}
		model.ParseSlackAttachment(post, []*model.SlackAttachment{bot.getConnectedAttachment(userLogin)})
		bot.DMUpdatePost(post)
	}

	return bot.Start(userID)
}

func (bot *mscBot) AfterDisconnect(userID string) error {
	errCancel := bot.Cancel(userID)
	errClean := bot.cleanPostIDs(userID)
	if errCancel != nil {
		return errCancel
	}

	if errClean != nil {
		return errClean
	}
	return nil
}

func (bot *mscBot) notifyWelcome(userID string) error {
	user, err := bot.Store.LoadUser(userID)
	if err != nil {
		return err
	}
	_, err = bot.DM(userID, WelcomeMessage, user.Remote.Mail)
	return err
}

func (bot *mscBot) WelcomeFlowEnd(userID string) {
	bot.notifySettings(userID)
}

func (bot *mscBot) getConnectAttachment() *model.SlackAttachment {
	sa := model.SlackAttachment{
		Title: "Connect",
		Text:  fmt.Sprintf(ConnectSuccessTemplate, bot.pluginURL),
	}

	return &sa
}

func (bot *mscBot) getConnectedAttachment(userLogin string) *model.SlackAttachment {
	return &model.SlackAttachment{
		Title: "Connect",
		Text:  ":tada: Congratulations! Your microsoft account (*" + userLogin + "*) has been connected to Mattermost.",
	}
}

func (bot *mscBot) notifySettings(userID string) error {
	_, err := bot.DM(userID, "Feel free to change these settings anytime") //[settings](%s/settings) anytime.", bot.pluginURL)
	if err != nil {
		return err
	}
	return nil
}

func (bot *mscBot) cleanPostIDs(mattermostUserID string) error {
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

func (bot *mscBot) SetProperty(userID, propertyName string, value bool) error {
	if propertyName != store.SubscribePropertyName {
		return bot.Dependencies.Store.SetProperty(userID, propertyName, value)
	}

	mscalendar := New(bot.Env, userID)
	_, err := mscalendar.CreateMyEventSubscription()
	if err != nil {
		return err
	}

	return nil
}

func (bot *mscBot) SetPostID(userID, propertyName, postID string) error {
	return bot.Dependencies.Store.SetPostID(userID, propertyName, postID)
}

func (bot *mscBot) GetPostID(userID, propertyName string) (string, error) {
	return bot.Dependencies.Store.GetPostID(userID, propertyName)
}

func (bot *mscBot) RemovePostID(userID, propertyName string) error {
	return bot.Dependencies.Store.RemovePostID(userID, propertyName)
}

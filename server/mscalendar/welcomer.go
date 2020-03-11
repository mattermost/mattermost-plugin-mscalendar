package mscalendar

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-server/v5/model"
)

type Welcomer interface {
	Welcome(userID string) error
	AfterSuccessfullyConnect(userID, userLogin string) error
	AfterUpdateStatus(userID string, status bool) error
	AfterSetConfirmations(userID string, set bool) error
	AfterDisconnect(userID string) error
}

type MSCalendarBot struct {
	bot.Bot
	Welcomer
	Env
	pluginURL string
}

const WelcomeMessage = `### Welcome to the Microsoft Calendar plugin!
Here is some info to prove we got you logged in
- Name: %s
`

func GetMSCalendarBot(bot bot.Bot, env Env, pluginURL string) *MSCalendarBot {
	return &MSCalendarBot{
		Bot:       bot,
		Env:       env,
		pluginURL: pluginURL,
	}
}

func (bot *MSCalendarBot) Welcome(userID string) error {
	bot.cleanPostIDs(userID)

	postID, err := bot.DMWithAttachments(userID, bot.getConnectAttachment())
	if err != nil {
		return err
	}

	bot.Store.StoreUserWelcomePost(userID, postID)

	return nil
}

func (bot *MSCalendarBot) AfterSuccessfullyConnect(userID, userLogin string) error {
	user, err := bot.Store.LoadUser(userID)
	if err != nil {
		return err
	}

	postID, err := bot.Store.DeleteUserWelcomePost(userID)
	if postID != "" {
		post := &model.Post{
			Id: postID,
		}
		model.ParseSlackAttachment(post, []*model.SlackAttachment{bot.getConnectedAttachment(userLogin)})
		bot.DMUpdatePost(post)
	}

	if !user.Flags.WelcomeUpdateStatus {
		return bot.notifyUpdateStatus(userID)
	}

	if !user.Flags.WelcomeGetConfirmation {
		return bot.notifyGetConfirmation(userID)
	}

	err = bot.notifyWelcome(userID)
	if err != nil {
		return err
	}

	return nil
}

func (bot *MSCalendarBot) AfterUpdateStatus(userID string, status bool) error {
	user, err := bot.Store.LoadUser(userID)
	if err != nil {
		return err
	}

	if status && !user.Flags.WelcomeGetConfirmation {
		return bot.notifyGetConfirmation(userID)
	}

	return bot.notifySettings(userID)
}

func (bot *MSCalendarBot) AfterSetConfirmations(userID string, set bool) error {
	return bot.notifySettings(userID)
}

func (bot *MSCalendarBot) AfterDisconnect(userID string) error {
	return bot.cleanPostIDs(userID)
}

func (bot *MSCalendarBot) notifyWelcome(userID string) error {
	user, err := bot.Store.LoadUser(userID)
	if err != nil {
		return err
	}
	_, err = bot.DM(userID, WelcomeMessage, user.Remote.Mail)
	return err
}

func (bot *MSCalendarBot) getConnectAttachment() *model.SlackAttachment {
	sa := model.SlackAttachment{
		Title: "Connect",
		Text: fmt.Sprintf(`Welcome to the Microsoft Calendar Bot.
[Click here to link your account.](%s/oauth2/connect)`, bot.pluginURL),
	}

	return &sa
}

func (bot *MSCalendarBot) getConnectedAttachment(userLogin string) *model.SlackAttachment {
	return &model.SlackAttachment{
		Title: "Connect",
		Text:  ":tada: Congratulations! Your microsoft account (*" + userLogin + "*) has been connected to Mattermost.",
	}
}

func (bot *MSCalendarBot) getUpdateStatusAttachment() *model.SlackAttachment {
	actionYes := model.PostAction{
		Name: "Yes - Update my status",
		Integration: &model.PostActionIntegration{
			URL: bot.pluginURL + "/welcomeBot/updateStatus?update_status=true",
		},
	}

	actionNo := model.PostAction{
		Name: "No - Don't update my status",
		Integration: &model.PostActionIntegration{
			URL: bot.pluginURL + "/welcomeBot/updateStatus?update_status=false",
		},
	}

	sa := model.SlackAttachment{
		Title:   "Update Status",
		Text:    "Do you wish your Mattermost status to be automatically updated to be *Do not disturb* at the time of your Microsoft Calendar events?",
		Actions: []*model.PostAction{&actionYes, &actionNo},
	}

	return &sa
}

func (bot *MSCalendarBot) getConfirmationAttachment() *model.SlackAttachment {
	actionYes := model.PostAction{
		Name: "Yes - I will like to get confirmations",
		Integration: &model.PostActionIntegration{
			URL: bot.pluginURL + "/welcomeBot/setConfirmations?get_confirmation=true",
		},
	}

	actionNo := model.PostAction{
		Name: "No - Update my status automatically",
		Integration: &model.PostActionIntegration{
			URL: bot.pluginURL + "/welcomeBot/setConfirmations?get_confirmation=false",
		},
	}

	sa := model.SlackAttachment{
		Title:   "Confirm status change",
		Text:    "Do you want to receive confirmations before we update your status for each event?",
		Actions: []*model.PostAction{&actionYes, &actionNo},
	}

	return &sa
}

func (bot *MSCalendarBot) notifyUpdateStatus(userID string) error {
	postID, err := bot.DMWithAttachments(userID, bot.getUpdateStatusAttachment())
	if err != nil {
		return err
	}

	user, err := bot.Store.LoadUser(userID)
	if err != nil {
		return err
	}
	user.Flags.WelcomeUpdateStatusPostID = postID
	err = bot.Store.StoreUser(user)
	if err != nil {
		return err
	}

	return nil
}

func (bot *MSCalendarBot) notifyGetConfirmation(userID string) error {
	postID, err := bot.DMWithAttachments(userID, bot.getConfirmationAttachment())
	if err != nil {
		return err
	}

	user, err := bot.Store.LoadUser(userID)
	if err != nil {
		return err
	}
	user.Flags.WelcomeGetConfirmationPostID = postID
	err = bot.Store.StoreUser(user)
	if err != nil {
		return err
	}

	return nil
}

func (bot *MSCalendarBot) notifySettings(userID string) error {
	_, err := bot.DM(userID, "Feel free to change these settings anytime") //[settings](%s/settings) anytime.", bot.pluginURL)
	if err != nil {
		return err
	}
	return nil
}

func (bot *MSCalendarBot) cleanPostIDs(mattermostUserID string) error {
	postID, err := bot.Store.DeleteUserWelcomePost(mattermostUserID)
	if err != nil {
		return err
	}

	if postID != "" {
		err := bot.DeletePost(postID)
		if err != nil {
			bot.Errorf(err.Error())
		}
	}

	user, err := bot.Store.LoadUser(mattermostUserID)
	if err != nil {
		// User does not exist and therefore there is nothing to clean
		return nil
	}

	if user.Flags.WelcomeUpdateStatusPostID != "" {
		err := bot.DeletePost(user.Flags.WelcomeUpdateStatusPostID)
		if err != nil {
			bot.Errorf(err.Error())
		}
		user.Flags.WelcomeUpdateStatusPostID = ""
	}

	if user.Flags.WelcomeGetConfirmationPostID != "" {
		err := bot.DeletePost(user.Flags.WelcomeGetConfirmationPostID)
		if err != nil {
			bot.Errorf(err.Error())
		}
		user.Flags.WelcomeGetConfirmationPostID = ""
	}

	err = bot.Store.StoreUser(user)
	if err != nil {
		return err
	}

	return nil
}

package mscalendar

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
)

type Welcomer interface {
	Welcome(userID string) error
	AfterSuccessfullyConnect(userID, userLogin string) error
	AfterUpdateStatus(userID string, status bool) error
	AfterSetConfirmations(userID string, set bool) error
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
	_, err := bot.DM(userID, "Welcome to the Microsoft Calendar Bot.")
	if err != nil {
		return err
	}

	bot.CleanPostIDs(userID)

	postID, err := bot.DM(userID, "[Click here to link your account.](%s/oauth2/connect)",
		bot.pluginURL)
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

	postID, _ := bot.Store.LoadUserWelcomePost(userID)
	if postID != "" {
		err = bot.DMUpdate(postID, ":tada: Congratulations! Your Microsoft account (%s) has been connected to Mattermost.", userLogin)
		if err != nil {
			return err
		}
		err := bot.Store.DeleteUserWelcomePost(userID)
		if err != nil {
			return err
		}
	}

	if !user.Flags.WelcomeUpdateStatus {
		return bot.NotifyUpdateStatus(userID)
	}

	if !user.Flags.WelcomeGetConfirmation {
		return bot.NotifyGetConfirmation(userID)
	}

	err = bot.NotifyWelcome(userID)
	if err != nil {
		return err
	}

	return nil
}

func (bot *MSCalendarBot) NotifyUpdateStatus(userID string) error {
	message := "Do you wish your Mattermost status to be automatically updated to be *Do not disturb* at the time of your Microsoft Calendar events?\n"
	message += "[Yes - Update my status](%s/welcomeBot/updateStatus?update_status=true)\t"
	message += "[No - Don't update my status](%s/welcomeBot/updateStatus?update_status=false)"
	postID, err := bot.DM(userID, message, bot.pluginURL, bot.pluginURL)
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

func (bot *MSCalendarBot) NotifyGetConfirmation(userID string) error {
	message := "Do you want to receive confirmations before we update your status for each event?\n"
	message += "[Yes - I will like to get confirmations](%s/welcomeBot/setConfirmations?set=true)\t"
	message += "[No - Update my status automatically](%s/welcomeBot/setConfirmations?set=false)\t"
	postID, err := bot.DM(userID, message, bot.pluginURL, bot.pluginURL)
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

func (bot *MSCalendarBot) NotifyWelcome(userID string) error {
	user, err := bot.Store.LoadUser(userID)
	if err != nil {
		return err
	}
	_, err = bot.DM(userID, WelcomeMessage, user.Remote.Mail)
	return err
}

func (bot *MSCalendarBot) NotifySettings(userID string) error {
	_, err := bot.DM(userID, "Feel free to change these [settings](%s/settings) anytime.", bot.pluginURL)
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

	if user.Flags.WelcomeUpdateStatusPostID != "" {
		message := ":thumbsup: Got it! We won't update your status in Mattermost."
		if status {
			message = ":thumbsup: Got it! We'll automatically update your status in Mattermost."
		}
		err = bot.DMUpdate(user.Flags.WelcomeUpdateStatusPostID, message)
		if err != nil {
			return err
		}
		user.Flags.WelcomeUpdateStatusPostID = ""
		err := bot.Store.StoreUser(user)
		if err != nil {
			return err
		}
	}

	if status && !user.Flags.WelcomeGetConfirmation {
		return bot.NotifyGetConfirmation(userID)
	}

	return bot.NotifySettings(userID)
}

func (bot *MSCalendarBot) AfterSetConfirmations(userID string, set bool) error {
	user, err := bot.Store.LoadUser(userID)
	if err != nil {
		return err
	}

	if user.Flags.WelcomeGetConfirmationPostID != "" {
		message := "Cool, we'll also send you confirmations before updating your status."
		if !set {
			message = "Cool, we will automatically update your status."
		}
		err = bot.DMUpdate(user.Flags.WelcomeGetConfirmationPostID, message)
		if err != nil {
			return err
		}
		user.Flags.WelcomeGetConfirmationPostID = ""
		err := bot.Store.StoreUser(user)
		if err != nil {
			return err
		}
	}

	return bot.NotifySettings(userID)
}

func (bot *MSCalendarBot) CleanPostIDs(mattermostUserID string) error {
	err := bot.Store.DeleteUserWelcomePost(mattermostUserID)
	if err != nil {
		return err
	}

	user, err := bot.Store.LoadUser(mattermostUserID)
	if err != nil {
		// User does not exist and therefore there is nothing to clean
		return nil
	}

	if user.Flags.WelcomeUpdateStatusPostID != "" {
		err := bot.DMUpdate(user.Flags.WelcomeUpdateStatusPostID, "New connection detected. Please, use the latest message")
		if err != nil {
			bot.Errorf(err.Error())
		}
		user.Flags.WelcomeUpdateStatusPostID = ""
	}

	if user.Flags.WelcomeGetConfirmationPostID != "" {
		err := bot.DMUpdate(user.Flags.WelcomeGetConfirmationPostID, "New connection detected. Please, use the latest message")
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

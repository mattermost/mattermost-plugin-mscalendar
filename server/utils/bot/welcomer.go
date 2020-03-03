package bot

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type Welcomer interface {
	Welcome(userID string) error
	AfterSuccessfullyConnect(userID, userLogin string) error
	AfterUpdateStatus(userID string, status bool) error
	AfterSetConfirmations(userID string, set bool) error
}

func (bot *bot) Welcome(userID string) error {
	_, err := bot.dmAndGetPostID(userID, "Welcome to the Microsoft Calendar Bot.")
	if err != nil {
		return err
	}
	postID, err := bot.dmAndGetPostID(userID, "[Click here to link your account.](%s/oauth2/connect)",
		bot.pluginURL)
	if err != nil {
		return err
	}

	if err := bot.StoreUserWelcomePost(userID, postID); err != nil {
		return err
	}

	return nil
}

func (bot *bot) AfterSuccessfullyConnect(userID, userLogin string) error {
	postID, err := bot.LoadUserWelcomePost(userID)
	if err != nil {
		return err
	}

	err = bot.dmUpdate(postID, ":tada: Congratulations! Your Microsoft account (%s) has been connected to Mattermost.", userLogin)
	if err != nil {
		return err
	}

	message := "Do you wish your Mattermost status to be automatically updated to be *Do not disturb* at the time of your Microsoft Calendar events?\n"
	message += "[Yes - Update my status](%s/welcomeBot/updateStatus?update_status=true)\t"
	message += "[No - Don't update my status](%s/welcomeBot/updateStatus?update_status=false)"
	postID, err = bot.dmAndGetPostID(userID, message, bot.pluginURL, bot.pluginURL)
	if err != nil {
		return err
	}

	err = bot.StoreUserWelcomePost(userID, postID)
	if err != nil {
		return err
	}

	return nil
}

func (bot *bot) AfterUpdateStatus(userID string, status bool) error {
	postID, err := bot.LoadUserWelcomePost(userID)
	if err != nil {
		return err
	}

	message := ":thumbsup: Got it! We won't update your status in Mattermost."
	if status {
		message = ":thumbsup: Got it! We'll automatically update your status in Mattermost."
	}
	err = bot.dmUpdate(postID, message)
	if err != nil {
		return err
	}

	message = "Do you want to receive confirmations before we update your status for each event?\n"
	message += "[Yes - I will like to get confirmations](%s/welcomeBot/setConfirmations?set=true)\t"
	message += "[No - Update my status automatically](%s/welcomeBot/setConfirmations?set=false)\t"
	postID, err = bot.dmAndGetPostID(userID, message, bot.pluginURL, bot.pluginURL)
	if err != nil {
		return err
	}

	if err := bot.StoreUserWelcomePost(userID, postID); err != nil {
		return err
	}

	return nil
}

func (bot *bot) AfterSetConfirmations(userID string, set bool) error {
	postID, err := bot.LoadUserWelcomePost(userID)
	if err != nil {
		return err
	}

	err = bot.dmUpdate(postID, "Cool, we'll also send you confirmations before updating your status.")
	if err != nil {
		return err
	}

	_, err = bot.dmAndGetPostID(userID, "Feel free to change these [settings](%s/settings) anytime.", bot.pluginURL)
	if err != nil {
		return err
	}
	return nil
}

func (bot *bot) LoadUserWelcomePost(mattermostUserId string) (string, error) {
	var postID string
	data, appErr := bot.pluginAPI.KVGet(mattermostUserId)
	if appErr != nil {
		return "", errors.WithMessage(appErr, "failed plugin KVGet")
	}
	if data == nil {
		return "", errors.New("key not found")
	}
	err := json.Unmarshal(data, &postID)
	if err != nil {
		return "", err
	}
	return postID, nil
}

func (bot *bot) StoreUserWelcomePost(mattermostUserId, postId string) error {
	data, err := json.Marshal(postId)
	if err != nil {
		return err
	}
	appErr := bot.pluginAPI.KVSet(mattermostUserId, data)
	if appErr != nil {
		return errors.WithMessagef(appErr, "failed plugin KVSet %q", mattermostUserId)
	}
	return nil
}

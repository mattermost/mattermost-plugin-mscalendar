package mocks

import "github.com/mattermost/mattermost-plugin-msoffice/server/utils"

var _ utils.BotPoster = &MockBotPoster{}

type MockBotPoster struct {
	Err error
}

func (bot *MockBotPoster) PostDirect(userID, message, postType string) error {
	return bot.Err
}

func (bot *MockBotPoster) PostEphemeral(userID, channelId, message string) {
	// nop
}

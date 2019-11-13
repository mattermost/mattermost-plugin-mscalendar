package testhttp

import (
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
	"github.com/stretchr/testify/mock"
)

var (
	_ utils.BotPoster = &mockBotPoster{}
)

func newMockBotPoster(err error) *mockBotPoster {
	b := &mockBotPoster{}

	b.On("PostDirect", mock.Anything, mock.Anything, mock.Anything).Return(err)

	return b
}

type mockBotPoster struct {
	mock.Mock
}

func (bot *mockBotPoster) PostDirect(userID, message, postType string) error {
	args := bot.Called(userID, message, postType)
	return args.Error(0)
}

func (bot *mockBotPoster) PostEphemeral(userID, channelId, message string) {
	// nop
}

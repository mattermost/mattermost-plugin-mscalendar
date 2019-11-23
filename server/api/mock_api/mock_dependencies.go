// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mock_api

import (
	gomock "github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-plugin-msoffice/server/api"
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote/mock_remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/store/mock_store"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/bot/mock_bot"
)

func NewMockDependencies(ctrl *gomock.Controller) api.Dependencies {
	return api.Dependencies{
		UserStore:         mock_store.NewMockUserStore(ctrl),
		OAuth2StateStore:  mock_store.NewMockOAuth2StateStore(ctrl),
		SubscriptionStore: mock_store.NewMockSubscriptionStore(ctrl),
		Logger:            utils.NilLogger,
		Poster:            mock_bot.NewMockPoster(ctrl),
		Remote:            mock_remote.NewMockRemote(ctrl),
	}
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package utils

import (
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
)

type BotPoster interface {
	PostDirect(userID, message, postType string) error
	PostEphemeral(userID, channelId, message string)
}

type botPoster struct {
	config *config.Config
	API    plugin.API
}

func NewBotPoster(conf *config.Config, api plugin.API) BotPoster {
	return &botPoster{
		config: conf,
		API:    api,
	}
}

func (poster *botPoster) PostDirect(userID, message, postType string) error {
	channel, err := poster.API.GetDirectChannel(userID, poster.config.BotUserId)
	if err != nil {
		poster.API.LogInfo("Couldn't get bot's DM channel", "user_id", userID)
		return err
	}

	post := &model.Post{
		UserId:    poster.config.BotUserId,
		ChannelId: channel.Id,
		Message:   message,
		Type:      postType,
	}

	if _, err := poster.API.CreatePost(post); err != nil {
		mlog.Error(err.Error())
		return err
	}

	return nil
}

func (poster *botPoster) PostEphemeral(userId, channelId, message string) {
	post := &model.Post{
		UserId:    poster.config.BotUserId,
		ChannelId: channelId,
		Message:   message,
	}
	_ = poster.API.SendEphemeralPost(userId, post)
}

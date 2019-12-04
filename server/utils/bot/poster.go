// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package bot

import (
	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
)

type Poster interface {
	PostDirect(userID, message, postType string) error
	PostEphemeral(userID, channelId, message string)
}

type poster struct {
	API    plugin.API
	config *config.Config
}

// NewPoster creates a new bot poster.
func NewPoster(api plugin.API, conf *config.Config) Poster {
	return &poster{
		API:    api,
		config: conf,
	}
}

func (poster *poster) PostDirect(userID, message, postType string) error {
	channel, err := poster.API.GetDirectChannel(userID, poster.config.BotUserID)
	if err != nil {
		poster.API.LogInfo("Couldn't get bot's DM channel", "user_id", userID)
		return err
	}

	post := &model.Post{
		UserId:    poster.config.BotUserID,
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

func (poster *poster) PostEphemeral(userId, channelId, message string) {
	post := &model.Post{
		UserId:    poster.config.BotUserID,
		ChannelId: channelId,
		Message:   message,
	}
	_ = poster.API.SendEphemeralPost(userId, post)
}

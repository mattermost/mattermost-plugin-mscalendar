// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package bot

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
)

type Poster interface {
	PostDirectf(userID, format string, args ...interface{}) error
	PostDirectAttachments(userID string, attachments ...*model.SlackAttachment) error
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

func (poster *poster) PostDirectf(userID, format string, args ...interface{}) error {
	return poster.postDirect(userID, &model.Post{
		Message: fmt.Sprintf(format, args...),
	})
}

func (poster *poster) PostDirectAttachments(userID string, attachments ...*model.SlackAttachment) error {
	post := model.Post{}
	model.ParseSlackAttachment(&post, attachments)
	return poster.postDirect(userID, &post)
}

func (poster *poster) postDirect(userID string, post *model.Post) error {
	channel, err := poster.API.GetDirectChannel(userID, poster.config.BotUserID)
	if err != nil {
		poster.API.LogInfo("Couldn't get bot's DM channel", "user_id", userID)
		return err
	}
	post.ChannelId = channel.Id
	post.UserId = poster.config.BotUserID
	if _, err := poster.API.CreatePost(post); err != nil {
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

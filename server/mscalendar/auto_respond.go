// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-server/v5/model"
)

const DefaultAutoRespondMessage = "This user is currently in a meeting."

type AutoRespond interface {
	HandleBusyDM(post *model.Post) error
	SetUserAutoRespondMessage(userID string, message string) error
	OpenAutoRespondDialog(request model.OpenDialogRequest) error
}

func (m *mscalendar) HandleBusyDM(post *model.Post) error {
	channel, err := m.PluginAPI.GetMattermostChannel(post.ChannelId)
	if err != nil {
		return err
	}

	if channel.Type != model.CHANNEL_DIRECT {
		return nil
	}

	usersInChannel, err := m.PluginAPI.GetMattermostUsersInChannel(post.ChannelId, model.CHANNEL_SORT_BY_USERNAME, 0, 2)
	if err != nil {
		return err
	}

	var storedRecipient *store.User
	for _, u := range usersInChannel {
		storedUser, _ := m.Store.LoadUser(u.Id)
		if u.Id != post.UserId {
			storedRecipient = storedUser
			break
		}
	}

	if storedRecipient == nil || !storedRecipient.Settings.AutoRespond || len(storedRecipient.ActiveEvents) == 0 {
		return nil
	}

	recipientStatus, err := m.PluginAPI.GetMattermostUserStatus(storedRecipient.MattermostUserID)
	if err != nil {
		return err
	}
	if recipientStatus.Status == model.STATUS_ONLINE {
		return nil
	}

	message := storedRecipient.Settings.AutoRespondMessage
	if message == "" {
		message = DefaultAutoRespondMessage
	}

	m.Poster.Ephemeral(post.UserId, post.ChannelId, message)
	return nil
}

func (m *mscalendar) SetUserAutoRespondMessage(userID string, message string) error {
	return m.Store.SetSetting(userID, store.AutoRespondMessageSettingID, message)
}

func (m *mscalendar) OpenAutoRespondDialog(request model.OpenDialogRequest) error {
	return m.PluginAPI.OpenInteractiveDialog(request)
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

type Users interface {
	GetActingUser() *User
	GetTimezone(user *User) (string, error)
}

type User struct {
	MattermostUserID string
	*store.User
	MattermostUser *model.User
}

func NewUser(mattermostUserID string) *User {
	return &User{
		MattermostUserID: mattermostUserID,
	}
}

func (user *User) Clone() *User {
	clone := *user
	clone.User = user.User.Clone()
	return &clone
}

func (m *mscalendar) GetActingUser() *User {
	return m.actingUser
}

func (m *mscalendar) ExpandUser(user *User) error {
	if user.User == nil {
		storedUser, err := m.Store.LoadUser(user.MattermostUserID)
		if err != nil {
			return errors.Wrap(err, "User not connected")
		}
		user.User = storedUser
	}
	if user.MattermostUser == nil {
		mattermostUser, err := m.PluginAPI.GetMattermostUser(user.MattermostUserID)
		if err != nil {
			return err
		}
		user.MattermostUser = mattermostUser
	}
	return nil
}

func (m *mscalendar) GetTimezone(user *User) (string, error) {
	err := m.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return "", err
	}

	settings, err := m.client.GetMailboxSettings(user.Remote.ID)
	if err != nil {
		return "", err
	}
	return settings.TimeZone, nil
}

func (user *User) String() string {
	if user.MattermostUser != nil {
		return fmt.Sprintf("@%s", user.MattermostUser.Username)
	} else {
		return fmt.Sprintf("%s", user.MattermostUserID)
	}
}

func (user *User) Markdown() string {
	if user.MattermostUser != nil {
		return fmt.Sprintf("@%s", user.MattermostUser.Username)
	} else {
		return fmt.Sprintf("UserID: `%s`", user.MattermostUserID)
	}
}

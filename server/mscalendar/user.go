// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

type Users interface {
	GetActingUser() *User
	GetTimezone(user *User) (string, error)
	DisconnectUser(mattermostUserID string) error
	GetRemoteUser(mattermostUserID string) (*remote.User, error)
	IsAuthorizedAdmin(mattermostUserID string) (bool, error)
	RefreshOAuth2Token(mattermostUserID string) (*oauth2.Token, error)
	RefreshAllOAuth2Tokens() error
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

func (m *mscalendar) DisconnectUser(mattermostUserID string) error {
	return m.Store.DeleteUser(mattermostUserID)
}

func (m *mscalendar) GetRemoteUser(mattermostUserID string) (*remote.User, error) {
	storedUser, err := m.Store.LoadUser(mattermostUserID)
	if err != nil {
		return nil, err
	}

	return storedUser.Remote, nil
}

func (m *mscalendar) IsAuthorizedAdmin(mattermostUserID string) (bool, error) {
	return m.Dependencies.IsAuthorizedAdmin(mattermostUserID)
}

func (m *mscalendar) RefreshOAuth2Token(mattermostUserID string) (*oauth2.Token, error) {
	return NewOAuth2App(m.Env).RefreshOAuth2Token(mattermostUserID)
}

func (m *mscalendar) RefreshAllOAuth2Tokens() error {
	userIndex, err := m.Store.LoadUserIndex()
	if err != nil {
		return err
	}

	for _, ui := range userIndex {
		_, err := m.RefreshOAuth2Token(ui.MattermostUserID)
		if err == nil {
			return err
		}
	}

	return nil
}

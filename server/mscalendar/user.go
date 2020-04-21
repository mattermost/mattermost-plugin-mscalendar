// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

type Users interface {
	GetActingUser() *User
	GetTimezone(user *User) (string, error)
	DisconnectUser(mattermostUserID string) error
	GetRemoteUser(mattermostUserID string) (*remote.User, error)
	IsAuthorizedAdmin(mattermostUserID string) (bool, error)
	GetUserSettings(user *User) (*store.Settings, error)
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
	err := m.ExpandRemoteUser(user)
	if err != nil {
		return err
	}
	err = m.ExpandMattermostUser(user)
	if err != nil {
		return err
	}
	return nil
}

func (m *mscalendar) ExpandRemoteUser(user *User) error {
	if user.User == nil {
		storedUser, err := m.Store.LoadUser(user.MattermostUserID)
		if err != nil {
			return errors.Wrap(err, "It looks like your Mattermost account is not connected to a Microsoft account. Please connect your account using `/mscalendar connect`.")
		}
		user.User = storedUser
	}
	return nil
}

func (m *mscalendar) ExpandMattermostUser(user *User) error {
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
		withRemoteUser(user),
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

func (m *mscalendar) GetTimezoneByID(mattermostUserID string) (string, error) {
	return m.GetTimezone(NewUser(mattermostUserID))
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
	err := m.Filter(
		withClient,
	)
	if err != nil {
		return err
	}

	storedUser, err := m.Store.LoadUser(mattermostUserID)
	if err != nil {
		return err
	}

	eventSubscriptionID := storedUser.Settings.EventSubscriptionID
	if eventSubscriptionID != "" {
		err = m.Store.DeleteUserSubscription(storedUser, eventSubscriptionID)
		if err != nil && err != store.ErrNotFound {
			return errors.WithMessagef(err, "failed to delete subscription %s", eventSubscriptionID)
		}

		err = m.client.DeleteSubscription(eventSubscriptionID)
		if err != nil {
			m.Logger.Errorf("failed to delete remote subscription %s: %v", eventSubscriptionID, err)
		}
	}

	err = m.Store.DeleteUser(mattermostUserID)
	if err != nil {
		return err
	}

	if mattermostUserID == m.Config.BotUserID {
		return nil
	}

	err = m.Store.DeleteUserFromIndex(mattermostUserID)
	if err != nil {
		return err
	}

	err = m.Store.DeleteDailySummaryUserSettings(mattermostUserID)
	if err != nil {
		return err
	}

	return nil
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

func (m *mscalendar) GetUserSettings(user *User) (*store.Settings, error) {
	err := m.Filter(
		withUserExpanded(user),
	)
	if err != nil {
		return nil, err
	}

	return &user.Settings, nil
}

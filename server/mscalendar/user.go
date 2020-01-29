// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

type Users interface {
	GetActingUser() *User
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
func (mscalendar *mscalendar) GetActingUser() *User {
	return mscalendar.actingUser
}

func (mscalendar *mscalendar) ExpandUser(user *User) error {
	if user.User == nil {
		storedUser, err := mscalendar.UserStore.LoadUser(user.MattermostUserID)
		if err != nil {
			return err
		}
		user.User = storedUser
	}
	if user.MattermostUser == nil {
		mattermostUser, err := mscalendar.PluginAPI.GetMattermostUser(user.MattermostUserID)
		if err != nil {
			return err
		}
		user.MattermostUser = mattermostUser
	}
	return nil
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

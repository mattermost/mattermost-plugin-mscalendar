// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package plugin_api

import (
	"github.com/mattermost/mattermost-server/v5/model"
)

type PluginAPI interface {
	GetUserStatus(userID string) (*model.Status, *model.AppError)
	GetUserStatusesByIds(userIDs []string) ([]*model.Status, *model.AppError)
	UpdateUserStatus(userID, status string) (*model.Status, *model.AppError)
	GetUser(userID string) (*model.User, *model.AppError)
}

package api

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type PluginAPIImpl struct {
	pluginAPI plugin.API
}

func (impl *PluginAPIImpl) GetUserStatus(userID string) (*model.Status, *model.AppError) {
	return impl.pluginAPI.GetUserStatus(userID)
}

func (impl *PluginAPIImpl) GetUserStatusesByIds(userIDs []string) ([]*model.Status, *model.AppError) {
	return impl.pluginAPI.GetUserStatusesByIds(userIDs)
}

func (impl *PluginAPIImpl) UpdateUserStatus(userID, status string) (*model.Status, *model.AppError) {
	return impl.pluginAPI.UpdateUserStatus(userID, status)
}

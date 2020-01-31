package api

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

func (api *api) GetUserTimezone(mattermostUserID string) (string, error) {
	client, err := api.MakeClient()
	if err != nil {
		return "", err
	}

	remoteUser, err := api.GetRemoteUser(mattermostUserID)
	if err != nil {
		return "", err
	}

	settings, err := client.GetUserMailboxSettings(remoteUser.ID)
	if err != nil {
		return "", err
	}

	return settings.TimeZone, nil
}

func (api *api) GetRemoteUser(mattermostUserID string) (*remote.User, error) {
	if api.user != nil && api.user.MattermostUserID == mattermostUserID {
		return api.user.Remote, nil
	}

	u, storeErr := api.UserStore.LoadUser(mattermostUserID)
	if storeErr != nil {
		return nil, storeErr
	}
	return u.Remote, nil
}

func (api *api) GetMattermostUser(mattermostUserID string) (*model.User, error) {
	u, err := api.PluginAPI.GetUser(mattermostUserID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get MM user")
	}

	return u, nil
}

func (api *api) DisconnectUser(mattermostUserID string) error {
	return api.Dependencies.UserStore.DeleteUser(mattermostUserID)
}

func (api *api) DisconnectBot() error {
	return api.Dependencies.UserStore.DeleteUser(api.BotUserID)
}

func (api *api) IsAuthorizedAdmin(mattermostUserID string) (bool, error) {
	return api.Dependencies.IsAuthorizedAdmin(mattermostUserID)
}

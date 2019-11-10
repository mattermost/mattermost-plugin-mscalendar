// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package bot

import (
	"io/ioutil"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/pkg/errors"
)

func EnsureWithProfileImage(api plugin.API, h plugin.Helpers, username, displayName, description, imagePath string) (string, error) {
	botUserID, err := h.EnsureBot(&model.Bot{
		Username:    username,
		DisplayName: displayName,
		Description: description,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to ensure bot account")
	}

	profileImage, err := ioutil.ReadFile(imagePath)
	if err != nil {
		return "", errors.Wrap(err, "couldn't read profile image")
	}
	if appErr := api.SetProfileImage(botUserID, profileImage); appErr != nil {
		return "", errors.Wrap(appErr, "couldn't set profile image")
	}

	return botUserID, nil
}

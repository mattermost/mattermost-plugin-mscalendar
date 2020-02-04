// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

const WelcomeMessage = `### Welcome to the Microsoft Calendar plugin!
Here is some info to prove we got you logged in
- Name: %s
`

const RemoteUserAlreadyConnected = "Remote user %s is already mapped to a MM user. Please run `/mscalendar disconnect` with account %s"
const RemoteUserAlreadyConnectedNotFound = "User %s is already mapped to a MM user, but the MM user could not be found."

func (api *api) InitOAuth2(mattermostUserID string) (url string, err error) {
	remoteUser, err := api.GetRemoteUser(mattermostUserID)
	if err == nil {
		return "", errors.Errorf("User is already connected to %s", remoteUser.Mail)
	}

	conf := api.Remote.NewOAuth2Config()
	state := fmt.Sprintf("%v_%v", model.NewId()[0:15], mattermostUserID)
	err = api.OAuth2StateStore.StoreOAuth2State(state)
	if err != nil {
		return "", err
	}

	return conf.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

func (api *api) CompleteOAuth2(authedUserID, code, state string) error {
	oconf := api.Remote.NewOAuth2Config()

	err := api.OAuth2StateStore.VerifyOAuth2State(state)
	if err != nil {
		return errors.WithMessage(err, "missing stored state")
	}

	mattermostUserID := strings.Split(state, "_")[1]

	if mattermostUserID != authedUserID {
		if mattermostUserID != api.Config.BotUserID {
			return errors.New("not authorized, user ID mismatch")
		}
		isAdmin, authErr := api.IsAuthorizedAdmin(authedUserID)
		if authErr != nil || !isAdmin {
			return errors.New("non-admin user attempting to set up bot account")
		}
	}

	ctx := context.Background()
	tok, err := oconf.Exchange(ctx, code)
	if err != nil {
		return err
	}

	client := api.Remote.MakeClient(ctx, tok)
	me, err := client.GetMe()
	if err != nil {
		return err
	}

	u := &store.User{
		PluginVersion:    api.Config.PluginVersion,
		MattermostUserID: mattermostUserID,
		Remote:           me,
		OAuth2Token:      tok,
	}

	uid, err := api.UserStore.LoadMattermostUserId(me.ID)
	if err == nil {
		user, userErr := api.GetMattermostUser(uid)
		if userErr == nil {
			api.Poster.DM(authedUserID, RemoteUserAlreadyConnected, me.Mail, user.Username)
			return errors.Errorf(RemoteUserAlreadyConnected, me.Mail, user.Username)
		} else {
			// Orphaned connected account. Let it be overwritten by passing through here?
			api.Poster.DM(authedUserID, RemoteUserAlreadyConnectedNotFound, me.Mail)
		}
	}

	err = api.UserStore.StoreUser(u)
	if err != nil {
		return err
	}

	if mattermostUserID == api.Config.BotUserID {
		api.Poster.DM(authedUserID, "Bot user connected to account %s.", me.Mail)
	} else {
		api.Poster.DM(mattermostUserID, WelcomeMessage, me.DisplayName)
	}

	return nil
}

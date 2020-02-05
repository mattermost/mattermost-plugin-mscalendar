// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/oauth2connect"
)

const WelcomeMessage = `### Welcome to the Microsoft Calendar plugin!
Here is some info to prove we got you logged in
- Name: %s
`

const RemoteUserAlreadyConnected = "Remote user %s is already mapped to a MM user. Please run `/mscalendar disconnect` with account %s"
const RemoteUserAlreadyConnectedNotFound = "User %s is already mapped to a MM user, but the MM user could not be found."

type oauth2App struct {
	Env
}

func NewOAuth2App(env Env) oauth2connect.App {
	return &oauth2App{
		Env: env,
	}
}

func (app *oauth2App) InitOAuth2(mattermostUserID string) (url string, err error) {
	user, err := app.Store.LoadUser(mattermostUserID)
	if err == nil {
		return "", errors.Errorf("User is already connected to %s", user.Remote.Mail)
	}

	conf := app.Remote.NewOAuth2Config()
	state := fmt.Sprintf("%v_%v", model.NewId()[0:15], mattermostUserID)
	err = app.Store.StoreOAuth2State(state)
	if err != nil {
		return "", err
	}

	return conf.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

func (app *oauth2App) CompleteOAuth2(authedUserID, code, state string) error {
	if authedUserID == "" || code == "" || state == "" {
		return errors.New("missing user, code or state")
	}

	oconf := app.Remote.NewOAuth2Config()

	err := app.Store.VerifyOAuth2State(state)
	if err != nil {
		return errors.WithMessage(err, "missing stored state")
	}

	mattermostUserID := strings.Split(state, "_")[1]
	if mattermostUserID != authedUserID {
		if mattermostUserID != app.Config.BotUserID {
			return errors.New("not authorized, user ID mismatch")
		}
		isAdmin, authErr := app.IsAuthorizedAdmin(authedUserID) // needs fix
		if authErr != nil || !isAdmin {
			return errors.New("non-admin user attempting to set up bot account")
		}
	}

	ctx := context.Background()
	tok, err := oconf.Exchange(ctx, code)
	if err != nil {
		return err
	}

	client := app.Remote.MakeClient(ctx, tok)
	me, err := client.GetMe()
	if err != nil {
		return err
	}

	uid, err := app.Store.LoadMattermostUserId(me.ID)
	if err == nil {
		user, userErr := app.PluginAPI.GetMattermostUser(uid)
		if userErr == nil {
			app.Poster.DM(authedUserID, RemoteUserAlreadyConnected, me.Mail, user.Username)
			return errors.Errorf(RemoteUserAlreadyConnected, me.Mail, user.Username)
		} else {
			// Orphaned connected account. Let it be overwritten by passing through here?
			app.Poster.DM(authedUserID, RemoteUserAlreadyConnectedNotFound, me.Mail)
		}
	}

	u := &store.User{
		PluginVersion:    app.Config.PluginVersion,
		MattermostUserID: mattermostUserID,
		Remote:           me,
		OAuth2Token:      tok,
	}

	err = app.Store.StoreUser(u)
	if err != nil {
		return err
	}

	if mattermostUserID == app.Config.BotUserID {
		app.Poster.DM(authedUserID, "Bot user connected to account %s.", me.Mail)
	} else {
		app.Poster.DM(mattermostUserID, WelcomeMessage, me.Mail)
	}

	return nil
}

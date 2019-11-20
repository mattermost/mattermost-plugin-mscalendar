// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-server/model"

	"github.com/mattermost/mattermost-plugin-msoffice/server/store"
)

func (api *api) InitOAuth2(userID string) (url string, err error) {
	conf := api.Remote.NewOAuth2Config()
	state := fmt.Sprintf("%v_%v", model.NewId()[0:15], userID)
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
		return errors.WithMessage(err, "missing stored state: "+err.Error())
	}

	mattermostUserID := strings.Split(state, "_")[1]
	if mattermostUserID != authedUserID {
		return errors.WithMessage(err, "not authorized")
	}

	ctx := context.Background()
	tok, err := oconf.Exchange(ctx, code)
	if err != nil {
		return err
	}

	client := api.Remote.NewClient(ctx, tok)
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

	err = api.UserStore.StoreUser(u)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("### Welcome to the Microsoft Office plugin!\n"+
		"Here is some info to prove we got you logged in\n"+
		"Name: %s \n", me.DisplayName)
	api.Poster.PostDirect(mattermostUserID, message, "custom_TODO")

	return nil
}

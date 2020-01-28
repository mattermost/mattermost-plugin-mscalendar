// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

//TODO move this into OAuth2 package, consolidate dependencies into an interface, then ->utils/oauth2

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

func (mscalendar *mscalendar) InitOAuth2(userID string) (url string, err error) {
	conf := mscalendar.Remote.NewOAuth2Config()
	state := fmt.Sprintf("%v_%v", model.NewId()[0:15], userID)
	err = mscalendar.OAuth2StateStore.StoreOAuth2State(state)
	if err != nil {
		return "", err
	}

	return conf.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

func (mscalendar *mscalendar) CompleteOAuth2(authedUserID, code, state string) error {
	oconf := mscalendar.Remote.NewOAuth2Config()

	err := mscalendar.OAuth2StateStore.VerifyOAuth2State(state)
	if err != nil {
		return errors.WithMessage(err, "missing stored state")
	}

	mattermostUserID := strings.Split(state, "_")[1]
	if mattermostUserID != authedUserID {
		return errors.New("not authorized, user ID mismatch")
	}

	ctx := context.Background()
	tok, err := oconf.Exchange(ctx, code)
	if err != nil {
		return err
	}

	client := mscalendar.Remote.MakeClient(ctx, tok)
	me, err := client.GetMe()
	if err != nil {
		return err
	}

	u := &store.User{
		PluginVersion:    mscalendar.Config.PluginVersion,
		MattermostUserID: mattermostUserID,
		Remote:           me,
		OAuth2Token:      tok,
	}

	err = mscalendar.UserStore.StoreUser(u)
	if err != nil {
		return err
	}

	mscalendar.Poster.DM(mattermostUserID, WelcomeMessage, me.DisplayName)

	return nil
}

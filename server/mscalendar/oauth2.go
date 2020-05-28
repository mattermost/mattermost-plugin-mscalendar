// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

const BotWelcomeMessage = "Bot user connected to account %s."

const RemoteUserAlreadyConnected = "%s account `%s` is already mapped to Mattermost account `%s`. Please run `/%s disconnect`, while logged in as the Mattermost account."
const RemoteUserAlreadyConnectedNotFound = "%s account `%s` is already mapped to a Mattermost account, but the Mattermost user could not be found."

func (mscal *mscalendar) OnCompleteOAuth2(mattermostUserID string, tok *oauth2.Token) {
	ctx := context.Background()

	client := mscal.Remote.MakeClient(ctx, tok)
	me, err := client.GetMe()
	if err != nil {
		mscal.Logger.Errorf(err.Error())
		return
	}

	uid, err := mscal.Store.LoadMattermostUserID(me.ID)
	if err == nil {
		user, userErr := mscal.PluginAPI.GetMattermostUser(uid)
		if userErr == nil {
			mscal.Poster.DM(mattermostUserID, RemoteUserAlreadyConnected, config.ApplicationName, me.Mail, config.CommandTrigger, user.Username)
			mscal.Logger.Errorf(RemoteUserAlreadyConnected, config.ApplicationName, me.Mail, config.CommandTrigger, user.Username)
			return
		} else {
			// Couldn't fetch connected MM account. Reject connect attempt.
			mscal.Poster.DM(mattermostUserID, RemoteUserAlreadyConnectedNotFound, config.ApplicationName, me.Mail)
			mscal.Logger.Errorf(RemoteUserAlreadyConnectedNotFound, config.ApplicationName, me.Mail)
			return
		}
	}

	u := &store.User{
		PluginVersion:    mscal.Config.PluginVersion,
		MattermostUserID: mattermostUserID,
		Remote:           me,
	}

	mailboxSettings, err := client.GetMailboxSettings(me.ID)
	if err != nil {
		mscal.Logger.Errorf("Cannot get mailbox settings.")
		return
	}

	u.Settings.DailySummary = &store.DailySummaryUserSettings{
		PostTime: "8:00AM",
		Timezone: mailboxSettings.TimeZone,
		Enable:   false,
	}

	err = mscal.Store.StoreUser(u)
	if err != nil {
		mscal.Logger.Errorf(err.Error())
		return
	}

	err = mscal.Store.StoreUserInIndex(u)
	if err != nil {
		mscal.Logger.Errorf(err.Error())
		return
	}

	mscal.Welcomer.AfterSuccessfullyConnect(mattermostUserID, me.Mail)
}

func (mscal *mscalendar) NewOAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     mscal.OAuth2ClientID,
		ClientSecret: mscal.OAuth2ClientSecret,
		Scopes: []string{
			"offline_access",
			"User.Read",
			"Calendars.ReadWrite",
			"Calendars.ReadWrite.Shared",
			"Mail.Read",
			"Mail.Send",
		},
		Endpoint: microsoft.AzureADEndpoint(mscal.OAuth2Authority),
	}
}

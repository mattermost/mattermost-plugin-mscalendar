// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
)

func (c *Command) adminConsent(parameters ...string) (string, bool, error) {
	isAdmin, err := c.MSCalendar.IsAuthorizedAdmin(c.Args.UserId)
	if !isAdmin || err != nil {
		return "Please have your administrator set up your application's permissions.", false, nil
	}

	_, err = c.MSCalendar.GetRemoteUser(c.Args.UserId)
	if err != nil {
		out := fmt.Sprintf("Please connect your %s account using `/%s connect`.", config.CommandTrigger, config.ApplicationName)
		return out, false, nil
	}

	token, err := c.MSCalendar.CreateNewAdminConsentToken(c.Args.UserId)
	if err != nil {
		return "Error creating token", false, err
	}

	tenantID := c.Config.OAuth2Authority
	clientID := c.Config.OAuth2ClientID
	redirectURI := fmt.Sprintf("%s%s/adminconsent", c.Config.PluginURL, config.PathAPI)
	remoteHost := "https://login.microsoftonline.com"
	link := fmt.Sprintf("%s/%s/adminconsent?client_id=%s&state=%s&redirect_uri=%s", remoteHost, tenantID, clientID, token, redirectURI)

	out := fmt.Sprintf("Click [here](%s) to grant admin consent to the application's permissions.", link)
	return out, false, nil
}

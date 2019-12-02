// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
)

// LoadSubscriptionCreatorF is a function that is supposed to verify that the
// subscription exists, and matches the user record. Returns user data and the
// stored subscription.

type Remote interface {
	NewClient(context.Context, *oauth2.Token) Client
	NewOAuth2Config() *oauth2.Config
	HandleNotification(http.ResponseWriter, *http.Request) []*Notification
}

var Makers = map[string]func(*config.Config, utils.Logger) Remote{}

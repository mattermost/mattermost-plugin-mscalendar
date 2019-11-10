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

type LoadSubscriptionCreatorF func(subscriptionID string) (*User, *oauth2.Token, string, *Subscription, error)

type Remote interface {
	NewClient(context.Context, *oauth2.Token) Client
	NewOAuth2Config() *oauth2.Config
	HandleEventNotification(http.ResponseWriter, *http.Request, LoadSubscriptionCreatorF) []*EventNotification
}

var Makers = map[string]func(*config.Config, utils.Logger) Remote{}

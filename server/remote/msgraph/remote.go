// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"

	msgraph "github.com/yaegashi/msgraph.go/v1.0"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
)

const Kind = "msgraph"

type impl struct{}

func init() {
	remote.Known[Kind] = &impl{}
}

// NewMicrosoftGraphClient creates a new client.
func (r *impl) NewClient(ctx context.Context, conf *config.Config, token *oauth2.Token, logger utils.Logger) remote.Client {
	httpClient := r.NewOAuth2Config(conf).Client(ctx, token)
	c := &client{
		ctx:        ctx,
		httpClient: httpClient,
		Logger:     logger,
		rbuilder:   msgraph.NewClient(httpClient),
	}
	return c
}

func (r *impl) NewOAuth2Config(conf *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     conf.OAuth2ClientId,
		ClientSecret: conf.OAuth2ClientSecret,
		RedirectURL:  conf.PluginURL + config.OAuth2Path + config.OAuth2CompletePath,
		Scopes: []string{
			"offline_access",
			"User.Read",
			"Calendars.ReadWrite",
			"Calendars.ReadWrite.Shared",
		},
		Endpoint: microsoft.AzureADEndpoint(conf.OAuth2Authority),
	}
}

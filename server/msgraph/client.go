// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"fmt"

	graph "github.com/jkrecek/msgraph-go"
	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
)

const OAuth2RedirectPath = "/oauth2/complete"

const (
	authURLEndpoint  = "https://login.microsoftonline.com/%s/oauth2/v2.0/authorize"
	tokenURLEndpoint = "https://login.microsoftonline.com/%s/oauth2/v2.0/token"
)

type Client interface {
	GetMe() (*graph.Me, error)
	GetUserCalendar(remoteUserId string) ([]*graph.Calendar, error)
}

type client struct {
	graph *graph.Client
}

// NewClient creates a new Client from a Token.
func NewClient(conf *config.Config, token *oauth2.Token) Client {
	c := &client{
		graph: graph.NewClient(GetOAuth2Config(conf), token),
	}
	return c
}

func GetOAuth2Config(conf *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     `c8b93992-c9e5-49d5-b49e-cde125ad2904`,
		ClientSecret: `7BIAb3QJ5HLEzJCJa4WVGKXBzbRH@@?:`,
		RedirectURL:  conf.PluginURL + OAuth2RedirectPath,
		Scopes: []string{
			"User.Read",
			"Calendars.ReadWrite",
			"Calendars.ReadWrite.Shared",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf(authURLEndpoint, "797c4cd6-c645-47e2-95e8-0ebd007d8119"),
			TokenURL: fmt.Sprintf(tokenURLEndpoint, "797c4cd6-c645-47e2-95e8-0ebd007d8119"),
		},
	}
}

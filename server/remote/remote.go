// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import (
	"context"

	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
)

type Remote interface {
	NewClient(context.Context, *config.Config, *oauth2.Token, utils.Logger) Client
	NewOAuth2Config(conf *config.Config) *oauth2.Config
	ParseEventWebhook(data []byte, conf *config.Config) ([]string, []*Event, error)
}

var Known = map[string]Remote{}

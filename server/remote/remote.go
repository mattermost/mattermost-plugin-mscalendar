// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import (
	"context"

	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
)

type Remote interface {
	NewClient(ctx context.Context, conf *config.Config, token *oauth2.Token) Client
	NewOAuth2Config(conf *config.Config) *oauth2.Config
}

var Known = map[string]Remote{}

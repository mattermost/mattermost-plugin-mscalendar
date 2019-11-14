// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

func (r *impl) ParseEventWebhook(data []byte, conf *config.Config) ([]string, []*remote.Event, error) {
	return nil, nil, nil
}

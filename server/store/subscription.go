// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

type Subscription struct {
	PluginVersion       string
	Remote              *remote.Subscription
	MattermostCreatorID string
}

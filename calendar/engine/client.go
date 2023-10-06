// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package engine

import (
	"context"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
)

type Client interface {
	MakeClient() (remote.Client, error)
	MakeSuperuserClient() (remote.Client, error)
}

func (m *mscalendar) MakeClient() (remote.Client, error) {
	err := m.Filter(withActingUserExpanded)
	if err != nil {
		return nil, err
	}

	return m.Remote.MakeClient(context.Background(), m.actingUser.OAuth2Token), nil
}

// REVIEW: google service account? maybe not needed. this is only used for the status sync batch requests
func (m *mscalendar) MakeSuperuserClient() (remote.Client, error) {
	return m.Remote.MakeSuperuserClient(context.Background())
}

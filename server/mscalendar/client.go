// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"context"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

type Client interface {
	MakeClient() (remote.Client, error)
	MakeSuperuserClient() remote.Client
}

func (m *mscalendar) MakeClient() (remote.Client, error) {
	err := m.Filter(withActingUserExpanded)
	if err != nil {
		return nil, err
	}

	return m.Remote.MakeClient(context.Background(), m.actingUser.OAuth2Token), nil
}

func (m *mscalendar) MakeSuperuserClient(client remote.Client) (remote.Client, error) {
	token, err := client.GetSuperuserToken()
	if err != nil {
		return nil, err
	}
	return m.Remote.MakeSuperuserClient(context.Background(), token), nil
}

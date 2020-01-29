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

func (mscalendar *mscalendar) MakeClient() (remote.Client, error) {
	err := mscalendar.Filter(withActingUserExpanded)
	if err != nil {
		return nil, err
	}

	return mscalendar.Remote.MakeClient(context.Background(), mscalendar.actingUser.OAuth2Token), nil
}

func (mscalendar *mscalendar) MakeSuperuserClient() remote.Client {
	return mscalendar.Remote.MakeSuperuserClient(context.Background())
}

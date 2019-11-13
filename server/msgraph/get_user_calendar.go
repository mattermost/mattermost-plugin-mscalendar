// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import graph "github.com/jkrecek/msgraph-go"

func (c *client) GetUserCalendar(remoteUserId string) ([]*graph.Calendar, error) {
	return c.graph.GetMeCalendar()
}

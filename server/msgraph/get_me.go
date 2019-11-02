// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import graph "github.com/jkrecek/msgraph-go"

func (c *client) GetMe() (*graph.Me, error) {
	return c.graph.GetMe()
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"

	msgraph "github.com/yaegashi/msgraph.go/v1.0"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

func (c *client) GetUserEvent(userID, eventID string) (*remote.Event, error) {
	e := &remote.Event{}

	err := c.rbuilder.Users().ID(userID).Events().ID(eventID).Request().JSONRequest(
		c.ctx, http.MethodGet, "", nil, &e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (c *client) AcceptUserEvent(userID, eventID string) error {
	dummy := &msgraph.EventAcceptRequestParameter{}
	return c.rbuilder.Users().ID(userID).Events().ID(eventID).Accept(dummy).Request().Post(c.ctx)
}

func (c *client) DeclineUserEvent(userID, eventID string) error {
	dummy := &msgraph.EventDeclineRequestParameter{}
	return c.rbuilder.Users().ID(userID).Events().ID(eventID).Decline(dummy).Request().Post(c.ctx)
}

func (c *client) TentativelyAcceptUserEvent(userID, eventID string) error {
	dummy := &msgraph.EventTentativelyAcceptRequestParameter{}
	return c.rbuilder.Users().ID(userID).Events().ID(eventID).TentativelyAccept(dummy).Request().Post(c.ctx)
}

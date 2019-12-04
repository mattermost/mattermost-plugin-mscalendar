// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"errors"
	"net/http"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

func (c *client) GetNotificationData(orig *remote.Notification) (*remote.Notification, error) {
	n := *orig
	wh := n.Webhook.(*webhook)
	switch wh.ResourceData.DataType {
	case "#Microsoft.Graph.Event":
		event := remote.Event{}
		_, err := c.Call(http.MethodGet, wh.Resource, nil, &event)
		if err != nil {
			c.LogDebug("Error fetching notification data resource",
				"URL", wh.Resource,
				"error", err.Error(),
				"subscriptionID", wh.SubscriptionID)
			return nil, err
		}
		n.Event = &event
		n.ChangeType = wh.ChangeType
		n.IsBare = false

	default:
		c.LogInfo("Unknown resource type: "+wh.ResourceData.DataType,
			"subscriptionID", wh.SubscriptionID)
		return nil, errors.New("Unknown resource type: " + wh.ResourceData.DataType)
	}

	return &n, nil
}

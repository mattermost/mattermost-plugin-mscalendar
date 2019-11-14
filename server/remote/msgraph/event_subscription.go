// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

func (c *client) CreateUserEventSubscription(userID, notificationURL string) (*remote.Subscription, error) {
	resource := "me/events"
	changeType := "created,updated,deleted"
	expirationDateTime := time.Now().Add(4230 * time.Minute)

	c.LogError("<><>", "NotificationURL", notificationURL)

	sub := &remote.Subscription{
		Resource:           resource,
		ChangeType:         changeType,
		ExpirationDateTime: expirationDateTime.Format(time.RFC3339),
		NotificationURL:    notificationURL,
	}
	err := c.rbuilder.Subscriptions().Request().JSONRequest(c.ctx, http.MethodPost, "", sub, sub)
	if err != nil {
		return nil, err
	}

	c.LogDebug("msgraph: created subscription", "userID", userID, "subscriptionID", sub.ID)
	return sub, nil
}

func (c *client) DeleteEventSubscription(subscriptionID string) error {
	err := c.rbuilder.Subscriptions().ID(subscriptionID).Request().Delete(c.ctx)
	if err != nil {
		return err
	}
	c.LogDebug("msgraph: deleted subscription", "subscriptionID", subscriptionID)
	return nil
}

func (c *client) RenewEventSubscription(subscriptionID string, expires time.Time) error {
	v := struct {
		ExpirationDateTime string `json:"expirationDateTime"`
	}{
		expires.Format(time.RFC3339),
	}
	err := c.rbuilder.Subscriptions().ID(subscriptionID).Request().JSONRequest(c.ctx, "PATCH", "", v, nil)
	if err != nil {
		return err
	}
	c.LogDebug("msgraph: renewed subscription until "+expires.Format(time.RFC3339), "subscriptionID", subscriptionID)
	return nil
}

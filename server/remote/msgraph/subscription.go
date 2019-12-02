// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

const subscribeTTL = 10 * time.Minute

func (c *client) CreateSubscription(notificationURL string) (*remote.Subscription, error) {
	sub := &remote.Subscription{
		Resource:           "me/events",
		ChangeType:         "created,updated,deleted",
		NotificationURL:    notificationURL,
		ExpirationDateTime: time.Now().Add(subscribeTTL).Format(time.RFC3339),
	}
	err := c.rbuilder.Subscriptions().Request().JSONRequest(c.ctx, http.MethodPost, "", sub, sub)
	if err != nil {
		return nil, err
	}
	c.LogDebug("msgraph: created subscription",
		"subscriptionID", sub.ID,
		"resource", sub.Resource,
		"changeType", sub.ChangeType,
		"expirationDateTime", sub.ExpirationDateTime)
	return sub, nil
}

func (c *client) DeleteSubscription(subscriptionID string) error {
	err := c.rbuilder.Subscriptions().ID(subscriptionID).Request().Delete(c.ctx)
	if err != nil {
		return err
	}
	c.LogDebug("msgraph: deleted subscription", "subscriptionID", subscriptionID)
	return nil
}

func (c *client) RenewSubscription(subscriptionID string) (*remote.Subscription, error) {
	expires := time.Now().Add(subscribeTTL)
	v := struct {
		ExpirationDateTime string `json:"expirationDateTime"`
	}{
		expires.Format(time.RFC3339),
	}
	sub := remote.Subscription{}
	err := c.rbuilder.Subscriptions().ID(subscriptionID).Request().JSONRequest(c.ctx, http.MethodPatch, "", v, &sub)
	if err != nil {
		return nil, err
	}
	c.LogDebug("msgraph: renewed subscription until "+sub.ExpirationDateTime, "subscriptionID", subscriptionID)
	return &sub, nil
}

func (c *client) ListSubscriptions() ([]*remote.Subscription, error) {
	var v struct {
		Value []*remote.Subscription `json:"value"`
	}
	err := c.rbuilder.Subscriptions().Request().JSONRequest(c.ctx, http.MethodGet, "", nil, &v)
	if err != nil {
		return nil, err
	}
	c.LogDebug(fmt.Sprintf("GetSubscriptions: returned %d subscriptions", len(v.Value)))
	return v.Value, nil
}

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

func (c *client) createSubscription(resource, changeType, notificationURL string) (*remote.Subscription, error) {
	sub := &remote.Subscription{
		Resource:           resource,
		ChangeType:         changeType,
		NotificationURL:    notificationURL,
		ExpirationDateTime: time.Now().Add(subscribeTTL).Format(time.RFC3339),
	}
	err := c.rbuilder.Subscriptions().Request().JSONRequest(c.ctx, http.MethodPost, "", sub, sub)
	if err != nil {
		return nil, err
	}
	c.LogDebug("msgraph: created subscription", "subscriptionID", sub.ID, "resource", sub.Resource, "changeType", sub.ChangeType)
	return sub, nil
}

func (c *client) CreateEventMessageSubscription(notificationURL string) (*remote.Subscription, error) {
	return c.createSubscription(
		// TODO make work: "me/messages?filter=microsoft.graph.eventMessage/meetingMessageType ne ''",
		"me/messages",
		"created",
		notificationURL,
	)
}

func (c *client) CreateEventSubscription(notificationURL string) (*remote.Subscription, error) {
	return c.createSubscription(
		"me/events",
		"created,updated,deleted",
		notificationURL,
	)
}

func (c *client) DeleteSubscription(subscriptionID string) error {
	err := c.rbuilder.Subscriptions().ID(subscriptionID).Request().Delete(c.ctx)
	if err != nil {
		return err
	}
	c.LogDebug("msgraph: deleted subscription", "subscriptionID", subscriptionID)
	return nil
}

func (c *client) RenewSubscription(subscriptionID string) (time.Time, error) {
	expires := time.Now().Add(subscribeTTL)
	v := struct {
		ExpirationDateTime string `json:"expirationDateTime"`
	}{
		expires.Format(time.RFC3339),
	}
	err := c.rbuilder.Subscriptions().ID(subscriptionID).Request().JSONRequest(c.ctx, http.MethodPatch, "", v, nil)
	if err != nil {
		return time.Time{}, err
	}
	c.LogDebug("msgraph: renewed subscription until "+v.ExpirationDateTime, "subscriptionID", subscriptionID)
	return expires, nil
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

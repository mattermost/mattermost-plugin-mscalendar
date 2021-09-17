// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package gcal

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
)

const subscribeTTL = 48 * time.Hour

const defaultCalendarName = "primary"
const googleSubscriptionType = "webhook"
const subscriptionSuffix = "_calendar_event_notifications"

func newRandomString() string {
	b := make([]byte, 96)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (c *client) CreateMySubscription(notificationURL, remoteUserID string) (*remote.Subscription, error) {
	service, err := calendar.NewService(context.Background(), option.WithHTTPClient(c.httpClient))
	if err != nil {
		return nil, errors.Wrap(err, "gcal CreateMySubscription, error creating service")
	}

	reqBody := &calendar.Channel{
		Id:      remoteUserID + subscriptionSuffix,
		Token:   "should_be_secret",
		Type:    googleSubscriptionType,
		Address: notificationURL,
		Params: map[string]string{
			"ttl": fmt.Sprintf("%d", int64(subscribeTTL.Seconds())),
		},
	}

	createSubscriptionRequest := service.Events.Watch(defaultCalendarName, reqBody)
	googleSubscription, err := createSubscriptionRequest.Do()
	if err != nil {
		return nil, errors.Wrap(err, "gcal CreateMySubscription, error creating subscription")
	}

	sub := &remote.Subscription{
		ID:       googleSubscription.Id,
		Resource: defaultCalendarName,
		// ChangeType:         "created,updated,deleted",
		NotificationURL:    notificationURL,
		ExpirationDateTime: time.Now().Add(time.Second * time.Duration(googleSubscription.Expiration)).Format(time.RFC3339),
		// ClientState:        newRandomString(),
		CreatorID: remoteUserID,
	}

	c.Logger.With(bot.LogContext{
		"subscriptionID": sub.ID,
		"resource":       sub.Resource,
		// "changeType":         sub.ChangeType,
		"expirationDateTime": sub.ExpirationDateTime,
	}).Debugf("gcal: created subscription.")

	return sub, nil
}

func (c *client) DeleteSubscription(subscriptionID string) error {
	service, err := calendar.NewService(context.Background(), option.WithHTTPClient(c.httpClient))
	if err != nil {
		return errors.Wrap(err, "gcal DeleteSubscription, error creating service")
	}

	stopRequest := service.Channels.Stop(&calendar.Channel{Id: subscriptionID})
	err = stopRequest.Do()

	if err != nil {
		return errors.Wrap(err, "gcal DeleteSubscription, error from google response")
	}

	c.Logger.With(bot.LogContext{
		"subscriptionID": subscriptionID,
	}).Debugf("gcal: deleted subscription.")

	return nil
}

func (c *client) RenewSubscription(notificationURL, remoteUserID, subscriptionID string) (*remote.Subscription, error) {
	err := c.DeleteSubscription(subscriptionID)
	if err != nil {
		return nil, errors.Wrap(err, "gcal RenewSubscription, error deleting subscription")
	}

	sub, err := c.CreateMySubscription(notificationURL, remoteUserID)
	if err != nil {
		return nil, errors.Wrap(err, "gcal RenewSubscription, error deleting subscription")
	}

	c.Logger.Debugf("gcal: renewed subscription.")

	return sub, nil
}

func (c *client) ListSubscriptions() ([]*remote.Subscription, error) {
	return nil, errors.New("gcal ListSubscriptions not implemented. only used for debug command")
}

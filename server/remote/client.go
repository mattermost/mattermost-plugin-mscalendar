// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import "time"

type Client interface {
	CreateUserEventSubscription(userID, notificationURL string) (*Subscription, error)
	RenewEventSubscription(subscriptionID string, expires time.Time) error
	DeleteEventSubscription(subscriptionID string) error
	GetMe() (*User, error)
	GetUserCalendars(userID string) ([]*Calendar, error)
	GetUserDefaultCalendarView(userID string, startTime, endTime time.Time) ([]*Event, error)
	GetUserEvent(userID, eventID string) (*Event, error)
}

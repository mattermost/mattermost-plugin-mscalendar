// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import "time"

type Client interface {
	Call(method, path string, in, out interface{}) (responseData []byte, err error)
	CreateEventSubscription(notificationURL string) (*Subscription, error)
	CreateEventMessageSubscription(notificationURL string) (*Subscription, error)
	ListSubscriptions() ([]*Subscription, error)
	RenewSubscription(subscriptionID string) (time.Time, error)
	DeleteSubscription(subscriptionID string) error
	GetMe() (*User, error)
	GetUserCalendars(userID string) ([]*Calendar, error)
	GetUserDefaultCalendarView(userID string, startTime, endTime time.Time) ([]*Event, error)
	GetUserEvent(userID, eventID string) (*Event, error)
}

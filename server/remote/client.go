// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import "time"

type Client interface {
	AcceptUserEvent(userID, eventID string) error
	Call(method, path string, in, out interface{}) (responseData []byte, err error)
	CreateSubscription(notificationURL string) (*Subscription, error)
	DeclineUserEvent(userID, eventID string) error
	DeleteSubscription(subscriptionID string) error
	CreateCalendar(calendarName string) (*Calendar, error)
	CreateEvent(calendarEvent *Event) (*Event, error)
	DeleteCalendarByID(calendarID string) error
	GetMe() (*User, error)
	GetNotificationData(*Notification) (*Notification, error)
	GetUserCalendars(userID string) ([]*Calendar, error)
	GetUserDefaultCalendarView(userID string, startTime, endTime time.Time) ([]*Event, error)
	GetUserEvent(userID, eventID string) (*Event, error)
	ListSubscriptions() ([]*Subscription, error)
	RenewSubscription(subscriptionID string) (*Subscription, error)
	TentativelyAcceptUserEvent(userID, eventID string) error
}

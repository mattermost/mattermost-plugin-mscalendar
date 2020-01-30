// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import (
	"net/url"
	"time"
)

type Client interface {
	AcceptUserEvent(userID, eventID string) error
	CallJSON(method, path string, in, out interface{}) (responseData []byte, err error)
	CallFormPost(method, path string, in url.Values, out interface{}) (responseData []byte, err error)
	CreateSubscription(notificationURL string) (*Subscription, error)
	DeclineUserEvent(userID, eventID string) error
	DeleteSubscription(subscriptionID string) error
	CreateCalendar(calendar *Calendar) (*Calendar, error)
	FindMeetingTimes(meetingParams *FindMeetingTimesParameters) (*MeetingTimeSuggestionResults, error)
	CreateEvent(calendarEvent *Event) (*Event, error)
	DeleteCalendar(calendarID string) error
	GetMe() (*User, error)
	GetNotificationData(*Notification) (*Notification, error)
	GetUserCalendars(userID string) ([]*Calendar, error)
	GetSchedule(remoteUserID string, schedules []string, startTime, endTime *DateTime, availabilityViewInterval int) ([]*ScheduleInformation, error)
	GetUserDefaultCalendarView(userID string, startTime, endTime time.Time) ([]*Event, error)
	GetUserEvent(userID, eventID string) (*Event, error)
	GetUserMailboxSettings(remoteUserID string) (*MailboxSettings, error)
	ListSubscriptions() ([]*Subscription, error)
	RenewSubscription(subscriptionID string) (*Subscription, error)
	TentativelyAcceptUserEvent(userID, eventID string) error
}

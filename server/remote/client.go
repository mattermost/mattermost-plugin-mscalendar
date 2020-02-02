// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import (
	"net/url"
	"time"
)

type Client interface {
	AcceptEvent(userID, eventID string) error
	CallFormPost(method, path string, in url.Values, out interface{}) (responseData []byte, err error)
	CallJSON(method, path string, in, out interface{}) (responseData []byte, err error)
	CreateCalendar(userID string, calendar *Calendar) (*Calendar, error)
	CreateEvent(userID string, calendarEvent *Event) (*Event, error)
	CreateMySubscription(notificationURL string) (*Subscription, error)
	DeclineEvent(userID, eventID string) error
	DeleteCalendar(userID, calendarID string) error
	DeleteSubscription(subscriptionID string) error
	FindMeetingTimes(userID string, meetingParams *FindMeetingTimesParameters) (*MeetingTimeSuggestionResults, error)
	GetCalendars(userID string) ([]*Calendar, error)
	GetDefaultCalendarView(userID string, startTime, endTime time.Time) ([]*Event, error)
	GetEvent(userID, eventID string) (*Event, error)
	GetMailboxSettings(userID string) (*MailboxSettings, error)
	GetMe() (*User, error)
	GetNotificationData(*Notification) (*Notification, error)
	GetSchedule(userID string, schedules []string, startTime, endTime *DateTime, availabilityViewInterval int) ([]*ScheduleInformation, error)
	ListSubscriptions() ([]*Subscription, error)
	RenewSubscription(subscriptionID string) (*Subscription, error)
	TentativelyAcceptEvent(userID, eventID string) error
}

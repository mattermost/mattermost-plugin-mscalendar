// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import (
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

type Client interface {
	Core
	Calendars
	Events
	Subscriptions
	Utils
	Unsupported
}

type Core interface {
	GetMe() (*User, error)
}

type myClientImplementation struct {
	token *oauth2.Token
}

// AcceptEvent implements Client.
func (c *myClientImplementation) AcceptEvent(remoteUserID string, eventID string) error {
	panic("unimplemented")
}

// CallFormPost implements Client.
func (c *myClientImplementation) CallFormPost(method string, path string, in url.Values, out interface{}) (responseData []byte, err error) {
	panic("unimplemented")
}

// CallJSON implements Client.
func (c *myClientImplementation) CallJSON(method string, path string, in interface{}, out interface{}) (responseData []byte, err error) {
	panic("unimplemented")
}

// CreateCalendar implements Client.
func (c *myClientImplementation) CreateCalendar(remoteUserID string, calendar *Calendar) (*Calendar, error) {
	panic("unimplemented")
}

// CreateEvent implements Client.
func (c *myClientImplementation) CreateEvent(remoteUserID string, calendarEvent *Event) (*Event, error) {
	panic("unimplemented")
}

// CreateMySubscription implements Client.
func (c *myClientImplementation) CreateMySubscription(notificationURL string, remoteUserID string) (*Subscription, error) {
	panic("unimplemented")
}

// DeclineEvent implements Client.
func (c *myClientImplementation) DeclineEvent(remoteUserID string, eventID string) error {
	panic("unimplemented")
}

// DeleteCalendar implements Client.
func (c *myClientImplementation) DeleteCalendar(remoteUserID string, calendarID string) error {
	panic("unimplemented")
}

// DeleteSubscription implements Client.
func (c *myClientImplementation) DeleteSubscription(sub *Subscription) error {
	panic("unimplemented")
}

// DoBatchViewCalendarRequests implements Client.
func (c *myClientImplementation) DoBatchViewCalendarRequests([]*ViewCalendarParams) ([]*ViewCalendarResponse, error) {
	panic("unimplemented")
}

// FindMeetingTimes implements Client.
func (c *myClientImplementation) FindMeetingTimes(remoteUserID string, meetingParams *FindMeetingTimesParameters) (*MeetingTimeSuggestionResults, error) {
	panic("unimplemented")
}

// GetCalendars implements Client.
func (c *myClientImplementation) GetCalendars(remoteUserID string) ([]*Calendar, error) {
	panic("unimplemented")
}

// GetDefaultCalendarView implements Client.
func (c *myClientImplementation) GetDefaultCalendarView(remoteUserID string, startTime time.Time, endTime time.Time) ([]*Event, error) {
	panic("unimplemented")
}

// GetEvent implements Client.
func (c *myClientImplementation) GetEvent(remoteUserID string, eventID string) (*Event, error) {
	panic("unimplemented")
}

// GetEventsBetweenDates implements Client.
func (c *myClientImplementation) GetEventsBetweenDates(remoteUserID string, start time.Time, end time.Time) ([]*Event, error) {
	panic("unimplemented")
}

// GetMailboxSettings implements Client.
func (c *myClientImplementation) GetMailboxSettings(remoteUserID string) (*MailboxSettings, error) {
	panic("unimplemented")
}

// GetNotificationData implements Client.
func (c *myClientImplementation) GetNotificationData(*Notification) (*Notification, error) {
	panic("unimplemented")
}

// GetSuperuserToken implements Client.
func (c *myClientImplementation) GetSuperuserToken() (string, error) {
	panic("unimplemented")
}

// ListSubscriptions implements Client.
func (c *myClientImplementation) ListSubscriptions() ([]*Subscription, error) {
	panic("unimplemented")
}

// RenewSubscription implements Client.
func (c *myClientImplementation) RenewSubscription(notificationURL string, remoteUserID string, sub *Subscription) (*Subscription, error) {
	panic("unimplemented")
}

// TentativelyAcceptEvent implements Client.
func (c *myClientImplementation) TentativelyAcceptEvent(remoteUserID string, eventID string) error {
	panic("unimplemented")
}

func NewClient(token *oauth2.Token) Client {
	return &myClientImplementation{token: token}
}

// Implement Client interface methods
func (c *myClientImplementation) GetMe() (*User, error) {
	// Example implementation
	return nil, nil
}

type Calendars interface {
	GetEvent(remoteUserID, eventID string) (*Event, error)
	GetCalendars(remoteUserID string) ([]*Calendar, error)
	GetDefaultCalendarView(remoteUserID string, startTime, endTime time.Time) ([]*Event, error)
	DoBatchViewCalendarRequests([]*ViewCalendarParams) ([]*ViewCalendarResponse, error)
	GetMailboxSettings(remoteUserID string) (*MailboxSettings, error)
}

type Events interface {
	CreateEvent(remoteUserID string, calendarEvent *Event) (*Event, error)
	AcceptEvent(remoteUserID, eventID string) error
	DeclineEvent(remoteUserID, eventID string) error
	TentativelyAcceptEvent(remoteUserID, eventID string) error
	GetEventsBetweenDates(remoteUserID string, start, end time.Time) ([]*Event, error)
}

type Subscriptions interface {
	CreateMySubscription(notificationURL, remoteUserID string) (*Subscription, error)
	DeleteSubscription(sub *Subscription) error
	GetNotificationData(*Notification) (*Notification, error)
	ListSubscriptions() ([]*Subscription, error)
	RenewSubscription(notificationURL, remoteUserID string, sub *Subscription) (*Subscription, error)
}

type Utils interface {
	GetSuperuserToken() (string, error)
	CallFormPost(method, path string, in url.Values, out interface{}) (responseData []byte, err error)
	CallJSON(method, path string, in, out interface{}) (responseData []byte, err error)
}

type Unsupported interface {
	CreateCalendar(remoteUserID string, calendar *Calendar) (*Calendar, error)
	DeleteCalendar(remoteUserID, calendarID string) error
	FindMeetingTimes(remoteUserID string, meetingParams *FindMeetingTimesParameters) (*MeetingTimeSuggestionResults, error)
}

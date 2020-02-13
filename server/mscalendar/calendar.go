// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/pkg/errors"
)

type Calendar interface {
	CreateCalendar(user *User, calendar *remote.Calendar) (*remote.Calendar, error)
	CreateEvent(user *User, event *remote.Event, mattermostUserIDs []string) (*remote.Event, error)
	DeleteCalendar(user *User, calendarID string) error
	FindMeetingTimes(user *User, meetingParams *remote.FindMeetingTimesParameters) (*remote.MeetingTimeSuggestionResults, error)
	GetCalendars(user *User) ([]*remote.Calendar, error)
	ViewCalendar(user *User, from, to time.Time) ([]*remote.Event, error)
	InitPolling(user *User) (events []*remote.Event, deltaURL string, err error)
	Poll(user *User) (events []*remote.Event, deltaURL string, err error)
}

func (m *mscalendar) ViewCalendar(user *User, from, to time.Time) ([]*remote.Event, error) {
	err := m.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return nil, err
	}
	return m.client.GetDefaultCalendarView(user.Remote.ID, from, to)
}

func (m *mscalendar) CreateCalendar(user *User, calendar *remote.Calendar) (*remote.Calendar, error) {
	err := m.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return nil, err
	}
	return m.client.CreateCalendar(user.Remote.ID, calendar)
}

func (m *mscalendar) CreateEvent(user *User, event *remote.Event, mattermostUserIDs []string) (*remote.Event, error) {
	err := m.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return nil, err
	}

	// invite non-mapped Mattermost
	for id := range mattermostUserIDs {
		mattermostUserID := mattermostUserIDs[id]
		_, err := m.Store.LoadUser(mattermostUserID)
		if err != nil {
			if err.Error() == "not found" {
				err = m.Poster.DM(mattermostUserID, "You have been invited to an MS office calendar event but have not linked your account.  Feel free to join us by connecting your www.office.com using `/mscalendar connect`")
			}
		}
	}

	return m.client.CreateEvent(user.Remote.ID, event)
}

func (m *mscalendar) DeleteCalendar(user *User, calendarID string) error {
	err := m.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return err
	}

	return m.client.DeleteCalendar(user.Remote.ID, calendarID)
}

func (m *mscalendar) FindMeetingTimes(user *User, meetingParams *remote.FindMeetingTimesParameters) (*remote.MeetingTimeSuggestionResults, error) {
	err := m.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return nil, err
	}

	return m.client.FindMeetingTimes(user.Remote.ID, meetingParams)
}

func (m *mscalendar) GetCalendars(user *User) ([]*remote.Calendar, error) {
	err := m.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return nil, err
	}

	return m.client.GetCalendars(user.Remote.ID)
}

// Poll uses the stored deltaLink in the user's subscription to fetch any events that have changed since the last fetch
// This should run frequently throughout the day to ensure we process new event updates, in case the Graph notifications are experiencing an outage
func (m *mscalendar) Poll(user *User) (events []*remote.Event, deltaURL string, err error) {
	err = m.Filter(
		withClient,
		withRemoteUser(user),
	)
	if err != nil {
		return nil, "", err
	}

	sub, err := m.loadUserSubscription(user.MattermostUserID)
	if err != nil {
		return
	}

	if sub.PollingURL == "" {
		return nil, "", errors.New("No polling URL stored for user " + user.Remote.Mail)
	}
	events, deltaURL, err = m.client.GetEventDeltaFromURL(sub.PollingURL)
	if err != nil {
		return nil, "", err
	}

	sub.PollingURL = deltaURL
	err = m.Store.StoreUserSubscription(user.User, sub)
	if err != nil {
		return
	}

	for _, event := range events {
		_, err := m.Store.LoadUserEvent(user.MattermostUserID, event.ID)
		if err != nil && err != store.ErrNotFound {
			m.Logger.Errorf("Failed to fetch event %s", err.Error())
			continue
		}

		changeType := "updated"
		if event.Body == nil {
			changeType = "deleted"
		} else if err == store.ErrNotFound {
			changeType = "created"
		}

		n := &remote.Notification{
			ChangeType:          changeType,
			SubscriptionID:      sub.Remote.ID,
			RecommendRenew:      false,
			ClientState:         sub.Remote.ClientState,
			Subscription:        sub.Remote,
			SubscriptionCreator: user.Remote,
			Event:               event,
		}

		m.Logger.With(bot.LogContext{
			"SubsriptionID": n.SubscriptionID,
			"ChangeType":    n.ChangeType,
			"EventID":       n.Event.ID,
		}).Debugf("Processing notification from delta")

		// TODO: delay enqueue
		// if notifications are working, event will already exist, and fields will be equal,
		// so this "notification" will be correctly dropped by notificationProcessor

		m.Env.NotificationProcessor.Enqueue(n)
	}

	return
}

// InitPolling is called when a user first subscribes to updates, and at the beginning of each day, for each user.
// It stores all events it receives, and the latest deltaLink in the user's subscription, to be used in subsequent calls to Poll.
func (m *mscalendar) InitPolling(user *User) (events []*remote.Event, deltaURL string, err error) {
	err = m.Filter(
		withClient,
		withRemoteUser(user),
	)
	if err != nil {
		return nil, "", err
	}

	start := time.Now().UTC()
	end := start.Add(time.Hour * time.Duration(24*30)).UTC()
	startDT := remote.NewDateTime(start, "UTC")
	endDT := remote.NewDateTime(end, "UTC")

	events, deltaURL, err = m.client.GetEventDeltaFromDateRange(user.Remote.ID, startDT, endDT)
	if err != nil {
		return nil, "", err
	}

	for _, event := range events {
		e := &store.Event{
			Remote:        event,
			PluginVersion: m.Config.PluginVersion,
		}
		err = m.Store.StoreUserEvent(user.MattermostUserID, e)
		if err != nil {
			return nil, "", err
		}
	}

	sub, err := m.loadUserSubscription(user.MattermostUserID)
	if err != nil {
		return
	}

	sub.PollingURL = deltaURL
	err = m.Store.StoreUserSubscription(user.User, sub)
	if err != nil {
		return
	}
	return
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

type Calendar interface {
	CreateCalendar(user *User, calendar *remote.Calendar) (*remote.Calendar, error)
	CreateEvent(user *User, event *remote.Event, mattermostUserIDs []string) (*remote.Event, error)
	DeleteCalendar(user *User, calendarID string) error
	FindMeetingTimes(user *User, meetingParams *remote.FindMeetingTimesParameters) (*remote.MeetingTimeSuggestionResults, error)
	GetCalendars(user *User) ([]*remote.Calendar, error)
	ViewCalendar(user *User, from, to time.Time) ([]*remote.Event, error)
}

func (mscalendar *mscalendar) ViewCalendar(user *User, from, to time.Time) ([]*remote.Event, error) {
	err := mscalendar.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return nil, err
	}
	return mscalendar.client.GetDefaultCalendarView(user.Remote.ID, from, to)
}

func (mscalendar *mscalendar) CreateCalendar(user *User, calendar *remote.Calendar) (*remote.Calendar, error) {
	err := mscalendar.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return nil, err
	}
	return mscalendar.client.CreateCalendar(user.Remote.ID, calendar)
}

func (mscalendar *mscalendar) CreateEvent(user *User, event *remote.Event, mattermostUserIDs []string) (*remote.Event, error) {
	err := mscalendar.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return nil, err
	}

	// invite non-mapped Mattermost
	for id := range mattermostUserIDs {
		userID := mattermostUserIDs[id]
		_, err := mscalendar.Store.LoadUser(userID)
		if err != nil {
			if err.Error() == "not found" {
				err = mscalendar.Poster.DM(userID, "You have been invited to an MS office calendar event but have not linked your account.  Feel free to join us by connecting your www.office.com using `/msoffice connect`")
			}
		}
	}

	return mscalendar.client.CreateEvent(user.Remote.ID, event)
}

func (mscalendar *mscalendar) DeleteCalendar(user *User, calendarID string) error {
	err := mscalendar.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return err
	}

	return mscalendar.client.DeleteCalendar(user.Remote.ID, calendarID)
}

func (mscalendar *mscalendar) FindMeetingTimes(user *User, meetingParams *remote.FindMeetingTimesParameters) (*remote.MeetingTimeSuggestionResults, error) {
	err := mscalendar.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return nil, err
	}

	return mscalendar.client.FindMeetingTimes(user.Remote.ID, meetingParams)
}

func (mscalendar *mscalendar) GetCalendars(user *User) ([]*remote.Calendar, error) {
	err := mscalendar.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return nil, err
	}

	return mscalendar.client.GetCalendars(user.Remote.ID)
}

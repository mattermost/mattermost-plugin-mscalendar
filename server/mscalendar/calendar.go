// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

type Calendar interface {
	ViewCalendar(from, to time.Time) ([]*remote.Event, error)
	CreateEvent(event *remote.Event, mattermostUserIDs []string) (*remote.Event, error)
	CreateCalendar(calendar *remote.Calendar) (*remote.Calendar, error)
	DeleteCalendar(calendarID string) error
	FindMeetingTimes(meetingParams *remote.FindMeetingTimesParameters) (*remote.MeetingTimeSuggestionResults, error)
	GetUserCalendars(userID string) ([]*remote.Calendar, error)
}

func (mscalendar *mscalendar) ViewCalendar(from, to time.Time) ([]*remote.Event, error) {
	client, err := mscalendar.MakeClient()
	if err != nil {
		return nil, err
	}

	return client.GetUserDefaultCalendarView(mscalendar.user.Remote.ID, from, to)
}

func (mscalendar *mscalendar) CreateCalendar(calendar *remote.Calendar) (*remote.Calendar, error) {
	client, err := mscalendar.MakeClient()
	if err != nil {
		return nil, err
	}

	return client.CreateCalendar(calendar)
}

func (mscalendar *mscalendar) CreateEvent(event *remote.Event, mattermostUserIDs []string) (*remote.Event, error) {

	// invite non-mapped Mattermost
	for id := range mattermostUserIDs {
		userID := mattermostUserIDs[id]
		_, err := mscalendar.UserStore.LoadUser(userID)
		if err != nil {
			if err.Error() == "not found" {
				err = mscalendar.Poster.DM(userID, "You have been invited to an MS office calendar event but have not linked your account.  Feel free to join us by connecting your www.office.com using `/msoffice connect`")
			}
		}
	}

	client, err := mscalendar.MakeClient()
	if err != nil {
		return nil, err
	}

	return client.CreateEvent(event)
}

func (mscalendar *mscalendar) DeleteCalendar(calendarID string) error {
	client, err := mscalendar.MakeClient()
	if err != nil {
		return err
	}

	return client.DeleteCalendar(calendarID)
}

func (mscalendar *mscalendar) FindMeetingTimes(meetingParams *remote.FindMeetingTimesParameters) (*remote.MeetingTimeSuggestionResults, error) {
	client, err := mscalendar.MakeClient()
	if err != nil {
		return nil, err
	}

	return client.FindMeetingTimes(meetingParams)
}

func (mscalendar *mscalendar) GetUserCalendars(userID string) ([]*remote.Calendar, error) {
	client, err := mscalendar.MakeClient()
	if err != nil {
		return nil, err
	}

	//DEBUG just use me as the user, for now
	me, err := client.GetMe()
	return client.GetUserCalendars(me.ID)
}

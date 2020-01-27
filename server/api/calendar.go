// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

func (api *api) ViewCalendar(from, to time.Time) ([]*remote.Event, error) {
	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}

	return client.GetUserDefaultCalendarView(api.user.Remote.ID, from, to)
}

func (api *api) CreateCalendar(calendar *remote.Calendar) (*remote.Calendar, error) {
	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}

	return client.CreateCalendar(calendar)
}

func (api *api) DeleteCalendar(calendarID string) error {
	client, err := api.MakeClient()
	if err != nil {
		return err
	}

	return client.DeleteCalendar(calendarID)
}

func (api *api) FindMeetingTimes(meetingParams *remote.FindMeetingTimesParameters) (*remote.MeetingTimeSuggestionResults, error) {
	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}

	return client.FindMeetingTimes(meetingParams)
}

func (api *api) GetUserCalendars(userID string) ([]*remote.Calendar, error) {
	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}

	//DEBUG just use me as the user, for now
	me, err := client.GetMe()
	return client.GetUserCalendars(me.ID)
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

func (api *api) ViewCalendar(from, to time.Time) ([]*remote.Event, error) {
	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}

	return client.GetUserDefaultCalendarView(api.user.Remote.ID, from, to)
}

func (api *api) CreateCalendar(calendarName string) (*remote.Calendar, error) {
	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}

	return client.CreateCalendar(calendarName)
}

func (api *api) CreateEvent(calendarEvent *remote.Event) (*remote.Event, error) {
	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}

	return client.CreateEvent(calendarEvent)
}

func (api *api) DeleteCalendarByID(calendarId string) error {
	client, err := api.MakeClient()
	if err != nil {
		return err
	}

	return client.DeleteCalendarByID(calendarId)
}

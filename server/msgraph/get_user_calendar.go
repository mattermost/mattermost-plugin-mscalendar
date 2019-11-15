// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import graph "github.com/jkrecek/msgraph-go"

func (c *client) GetUserCalendar(remoteUserId string) ([]*graph.Calendar, error) {
	return c.graph.GetMeCalendar()
}

func (c *client) CreateCalendarEvent(calendarId string, event *graph.Event) (*graph.Event, error) {
	return c.graph.CreateCalendarEvent(calendarId, event)
}

func (c *client) CreateCalendar(calendar *graph.Calendar) (*graph.Calendar, error) {
	return c.graph.CreateCalendar(calendar)
}

func (c *client) GetCalendarEvents(calendarId string) ([]*graph.Event, error) {
	return c.graph.GetCalendarEvents(calendarId)
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"fmt"
	"sort"
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
	RenderCalendarView(user *User, events []*remote.Event) (string, error)
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

func (m *mscalendar) viewTodayCalendar(user *User) ([]*remote.Event, error) {
	err := m.Filter(
		withClient,
	)
	if err != nil {
		return nil, err
	}

	err = m.ExpandRemoteUser(user)
	if err != nil {
		return nil, err
	}

	from := time.Now()
	to := time.Now().Add(time.Hour * 24)
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

func (m *mscalendar) RenderCalendarView(user *User, events []*remote.Event) (string, error) {
	if len(events) == 0 {
		return "No events were found", nil
	}

	err := m.Filter(
		withUserExpanded(user),
	)
	if err != nil {
		return "", err
	}

	var timeZone string
	tz, err := m.GetTimezone(user)
	if err == nil {
		timeZone = tz
	}

	if timeZone != "" {
		for _, e := range events {
			e.Start = e.Start.In(timeZone)
			e.End = e.End.In(timeZone)
		}
	}

	resp := "Times are shown in " + events[0].Start.TimeZone + "\n"
	for _, group := range groupEventsByDate(events) {
		resp += group[0].Start.Time().Format("Monday January 02") + "\n"
		for _, e := range group {
			resp += fmt.Sprintf("* %s\n", renderEvent(e))
		}
	}

	return resp, nil
}

func renderEvent(event *remote.Event) string {
	start := event.Start.Time().Format(time.Kitchen)
	end := event.End.Time().Format(time.Kitchen)

	return fmt.Sprintf("%s - %s `%s`", start, end, event.Subject)
}

func groupEventsByDate(events []*remote.Event) [][]*remote.Event {
	groups := map[string][]*remote.Event{}

	for _, event := range events {
		date := event.Start.Time().Format("2006-01-02")
		_, ok := groups[date]
		if !ok {
			groups[date] = []*remote.Event{}
		}

		groups[date] = append(groups[date], event)
	}

	days := []string{}
	for k := range groups {
		days = append(days, k)
	}
	sort.Strings(days)

	result := [][]*remote.Event{}
	for _, day := range days {
		group := groups[day]
		result = append(result, group)
	}
	return result
}

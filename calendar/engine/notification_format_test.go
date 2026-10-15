// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package engine

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine/views"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
)

func TestEventToFieldsUserTimeFormatPreference(t *testing.T) {
	start := time.Date(2026, 10, 11, 14, 30, 0, 0, time.UTC)
	end := time.Date(2026, 10, 11, 16, 0, 0, 0, time.UTC)
	event := &remote.Event{
		ID:      "event-1",
		Subject: "Sprint Retrospective",
		Start:   remote.NewDateTime(start, "UTC"),
		End:     remote.NewDateTime(end, "UTC"),
		Organizer: &remote.Attendee{
			EmailAddress: &remote.EmailAddress{
				Name:    "Alice",
				Address: "alice@example.com",
			},
		},
		ResponseStatus: &remote.EventResponseStatus{
			Response: "accepted",
		},
	}

	t.Run("default 12-hour format in eventToFields", func(t *testing.T) {
		f := eventToFields(event, "UTC")
		dateField := f[FieldWhen]
		require.NotNil(t, dateField)
		require.Contains(t, strings.Join(dateField.Strings(), ", "), "2:30 PM - 4:00 PM")
	})

	t.Run("explicit 24-hour military format in eventToFields", func(t *testing.T) {
		f := eventToFields(event, "UTC", views.TimeFormat24Hour)
		dateField := f[FieldWhen]
		require.NotNil(t, dateField)
		require.Contains(t, strings.Join(dateField.Strings(), ", "), "14:30 - 16:00")
	})

	t.Run("all day event in eventToFields", func(t *testing.T) {
		allDayEvent := &remote.Event{
			ID:       "event-allday",
			Subject:  "All Day Workshop",
			IsAllDay: true,
			Start:    remote.NewDateTime(start, "UTC"),
			End:      remote.NewDateTime(end, "UTC"),
			Organizer: &remote.Attendee{
				EmailAddress: &remote.EmailAddress{
					Name:    "Alice",
					Address: "alice@example.com",
				},
			},
			ResponseStatus: &remote.EventResponseStatus{
				Response: "accepted",
			},
		}
		f := eventToFields(allDayEvent, "UTC", views.TimeFormat24Hour)
		dateField := f[FieldWhen]
		require.NotNil(t, dateField)
		require.Contains(t, strings.Join(dateField.Strings(), ", "), "All day event")
	})
}

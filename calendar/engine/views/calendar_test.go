// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package views

import (
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
)

func TestMarkdownToHTMLEntities(t *testing.T) {
	for _, testCase := range []struct {
		description    string
		inputstring    string
		expectedOutput string
	}{
		{
			description:    "with asterisk",
			inputstring:    "**bold text**",
			expectedOutput: "&#42;&#42;bold text&#42;&#42;",
		},
		{
			description:    "normal string",
			inputstring:    "normal string",
			expectedOutput: "normal string",
		},
		{
			description:    "with braces",
			inputstring:    "[square](round)",
			expectedOutput: "&#91;square&#93;&#40;round&#41;",
		},
		{
			description:    "with underscore",
			inputstring:    "text_test",
			expectedOutput: "text&#95;test",
		},
		{
			description:    "withbacktick",
			inputstring:    "`test`",
			expectedOutput: "&#96;test&#96;",
		},
		{
			description:    "with greater and less than",
			inputstring:    "<test>",
			expectedOutput: "&#60;test&#62;",
		},
		{
			description:    "with backslash",
			inputstring:    "test \\ text",
			expectedOutput: "test &#92; text",
		},
		{
			description:    "URL 1",
			inputstring:    "www.example.com",
			expectedOutput: "www&#46;example&#46;com",
		},
		{
			description:    "URL 2",
			inputstring:    "https://example.com",
			expectedOutput: "https&#58;&#47;&#47;example&#46;com",
		},
		{
			description:    "strike through",
			inputstring:    "~~strike~~",
			expectedOutput: "&#126;&#126;strike&#126;&#126;",
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			res := MarkdownToHTMLEntities(testCase.inputstring)
			require.EqualValues(t, testCase.expectedOutput, res)
		})
	}
}

func TestRenderCalendarView(t *testing.T) {
	start := time.Date(2026, 10, 11, 14, 30, 0, 0, time.UTC)
	end := time.Date(2026, 10, 11, 16, 0, 0, 0, time.UTC)
	events := []*remote.Event{
		{
			ID:      "event-1",
			Subject: "Sprint Planning",
			Start:   remote.NewDateTime(start, "UTC"),
			End:     remote.NewDateTime(end, "UTC"),
		},
	}

	t.Run("empty events", func(t *testing.T) {
		res, err := RenderCalendarView(nil, "UTC")
		require.NoError(t, err)
		require.Equal(t, "You have no upcoming events.", res)
	})

	t.Run("default 12-hour time format", func(t *testing.T) {
		res, err := RenderCalendarView(events, "UTC")
		require.NoError(t, err)
		require.Contains(t, res, "Sprint Planning")
		require.Contains(t, res, "2:30 PM - 4:00 PM")
	})

	t.Run("explicit 24-hour military time format", func(t *testing.T) {
		res, err := RenderCalendarView(events, "UTC", TimeFormat24Hour)
		require.NoError(t, err)
		require.Contains(t, res, "Sprint Planning")
		require.Contains(t, res, "14:30 - 16:00")
	})

	t.Run("all day event", func(t *testing.T) {
		allDayEvents := []*remote.Event{
			{
				ID:       "event-2",
				Subject:  "Hackathon Day",
				IsAllDay: true,
				Start:    remote.NewDateTime(start, "UTC"),
				End:      remote.NewDateTime(end, "UTC"),
			},
		}
		res, err := RenderCalendarView(allDayEvents, "UTC", TimeFormat24Hour)
		require.NoError(t, err)
		require.Contains(t, res, "All day event")
		require.Contains(t, res, "Hackathon Day")
	})
}

func TestRenderDaySummary(t *testing.T) {
	start := time.Date(2026, 10, 11, 9, 15, 0, 0, time.UTC)
	end := time.Date(2026, 10, 11, 10, 45, 0, 0, time.UTC)
	events := []*remote.Event{
		{
			ID:      "event-summary",
			Subject: "Design Review",
			Start:   remote.NewDateTime(start, "UTC"),
			End:     remote.NewDateTime(end, "UTC"),
		},
	}

	t.Run("empty events", func(t *testing.T) {
		msg, atts, err := RenderDaySummary(nil, "UTC")
		require.NoError(t, err)
		require.Equal(t, "You have no events for that day", msg)
		require.Nil(t, atts)
	})

	t.Run("default 12-hour format", func(t *testing.T) {
		_, atts, err := RenderDaySummary(events, "UTC")
		require.NoError(t, err)
		require.Len(t, atts, 1)
		require.Contains(t, atts[0].Text, "9:15 AM - 10:45 AM")
	})

	t.Run("24-hour military format", func(t *testing.T) {
		_, atts, err := RenderDaySummary(events, "UTC", TimeFormat24Hour)
		require.NoError(t, err)
		require.Len(t, atts, 1)
		require.Contains(t, atts[0].Text, "09:15 - 10:45")
	})
}

func TestRenderUpcomingEvent(t *testing.T) {
	start := time.Date(2026, 10, 11, 17, 0, 0, 0, time.UTC)
	end := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC)
	event := &remote.Event{
		ID:      "event-up",
		Subject: "Architecture Sync",
		Start:   remote.NewDateTime(start, "UTC"),
		End:     remote.NewDateTime(end, "UTC"),
	}

	t.Run("default 12-hour format", func(t *testing.T) {
		res, err := RenderUpcomingEvent(event, "UTC")
		require.NoError(t, err)
		require.Contains(t, res, "5:00 PM - 6:00 PM")
	})

	t.Run("24-hour military format", func(t *testing.T) {
		res, err := RenderUpcomingEvent(event, "UTC", TimeFormat24Hour)
		require.NoError(t, err)
		require.Contains(t, res, "17:00 - 18:00")
	})
}

func TestAttachmentOptions(t *testing.T) {
	start := time.Date(2026, 10, 11, 13, 0, 0, 0, time.UTC)
	end := time.Date(2026, 10, 11, 14, 0, 0, 0, time.UTC)
	event := &remote.Event{
		ID:      "event-opts",
		Subject: "Product Demo",
		Start:   remote.NewDateTime(start, "UTC"),
		End:     remote.NewDateTime(end, "UTC"),
	}

	t.Run("TimeFormatOption with 24-hour format", func(t *testing.T) {
		att, err := RenderEventAsAttachment(event, "UTC", TimeFormatOption(TimeFormat24Hour, "UTC"))
		require.NoError(t, err)
		require.Equal(t, "13:00 - 14:00", att.Text)
		require.Contains(t, att.Fallback, "13:00 - 14:00")
	})

	t.Run("ShowTimezoneOption with 24-hour format", func(t *testing.T) {
		att, err := RenderEventAsAttachment(event, "UTC", ShowTimezoneOption("UTC", TimeFormat24Hour))
		require.NoError(t, err)
		require.Equal(t, "13:00 - 14:00 (UTC)", att.Text)
		require.Equal(t, "Product Demo\n13:00 - 14:00 (UTC)", att.Fallback)
	})

	t.Run("ShowTimezoneOption with 12-hour default", func(t *testing.T) {
		att, err := RenderEventAsAttachment(event, "UTC", ShowTimezoneOption("UTC"))
		require.NoError(t, err)
		require.Equal(t, "1:00 PM - 2:00 PM (UTC)", att.Text)
		require.Equal(t, "Product Demo\n1:00 PM - 2:00 PM (UTC)", att.Fallback)
	})
}

func TestEventTimeFormat(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Madrid")
	require.NoError(t, err)

	startTime := time.Date(2023, 8, 9, 18, 0, 0, 0, loc)
	endTime := time.Date(2023, 8, 9, 18, 15, 0, 0, loc)

	event := &remote.Event{
		Subject: "Team Sync",
		Start:   remote.NewDateTime(startTime, "Europe/Madrid"),
		End:     remote.NewDateTime(endTime, "Europe/Madrid"),
	}

	opt := ShowTimezoneOption("Europe/Madrid")
	attachment := &model.MessageAttachment{}
	opt.Apply(*event, attachment)

	require.Equal(t, "6:00 PM - 6:15 PM (Europe/Madrid)", attachment.Text)
}

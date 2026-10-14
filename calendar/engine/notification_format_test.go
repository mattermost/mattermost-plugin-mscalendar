// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package engine

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine/mock_plugin_api"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine/views"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot/mock_bot"
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

func TestNotificationProcessorGetUserTimeFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPluginAPI := mock_plugin_api.NewMockPluginAPI(ctrl)
	logger := mock_bot.NewMockLogger(ctrl)

	processor := &notificationProcessor{
		Env: Env{
			Dependencies: &Dependencies{
				PluginAPI: mockPluginAPI,
				Logger:    logger,
			},
		},
	}

	t.Run("military time enabled", func(t *testing.T) {
		mockPluginAPI.EXPECT().GetUserPreference("user_1", model.PreferenceCategoryDisplaySettings, model.PreferenceNameUseMilitaryTime).Return("true", nil).Times(1)
		tf := processor.getUserTimeFormat("user_1")
		require.Equal(t, views.TimeFormat24Hour, tf)
	})

	t.Run("military time disabled", func(t *testing.T) {
		mockPluginAPI.EXPECT().GetUserPreference("user_1", model.PreferenceCategoryDisplaySettings, model.PreferenceNameUseMilitaryTime).Return("false", nil).Times(1)
		tf := processor.getUserTimeFormat("user_1")
		require.Equal(t, views.TimeFormat12Hour, tf)
	})

	t.Run("error looking up preference defaults to 12-hour", func(t *testing.T) {
		mockPluginAPI.EXPECT().GetUserPreference("user_1", model.PreferenceCategoryDisplaySettings, model.PreferenceNameUseMilitaryTime).Return("", errors.New("lookup error")).Times(1)
		logger.EXPECT().With(gomock.Any()).Return(logger).Times(1)
		logger.EXPECT().Debugf(gomock.Any()).Times(1)

		tf := processor.getUserTimeFormat("user_1")
		require.Equal(t, views.TimeFormat12Hour, tf)
	})

	t.Run("nil PluginAPI defaults to 12-hour", func(t *testing.T) {
		emptyProcessor := &notificationProcessor{}
		tf := emptyProcessor.getUserTimeFormat("user_1")
		require.Equal(t, views.TimeFormat12Hour, tf)
	})
}

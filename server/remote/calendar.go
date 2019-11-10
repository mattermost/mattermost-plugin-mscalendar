// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

type Calendar struct {
	ID           string  `json:"id"`
	Name         string  `json:"name,omitempty"`
	Events       []Event `json:"events,omitempty"`
	CalendarView []Event `json:"calendarView,omitempty"`
}

type Event struct {
	ID            string `json:"id"`
	Subject       string `json:"subject,omitempty"`
	BodyPreview   string `json:"body_preview,omitempty"`
	Start         string `json:"start,omitempty"`
	StartTimeZone string `json:"start_timezone,omitempty"`
	End           string `json:"end,omitempty"`
	EndTimezone   string `json:"end_timezone,omitempty"`
	Location      string `json:"location,omitempty"`
	IsAllDay      bool   `json:"is_all_day,omitempty"`
	IsCancelled   bool   `json:"is_cancelled,omitempty"`
}

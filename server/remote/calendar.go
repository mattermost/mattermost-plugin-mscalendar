// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import (
	"time"
)

type Calendar struct {
	ID           string  `json:"id"`
	Name         string  `json:"name,omitempty"`
	Events       []Event `json:"events,omitempty"`
	CalendarView []Event `json:"calendarView,omitempty"`
	Owner        *User   `json:"owner,omitempty"`
}

type ViewCalendarParams struct {
	RemoteID  string
	StartTime time.Time
	EndTime   time.Time
}

type ViewCalendarResponse struct {
	RemoteID string
	Events   []*Event
	Error    *ApiError
}

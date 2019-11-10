// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import "time"

type Client interface {
	GetMe() (*User, error)
	GetUserCalendars(userID string) ([]*Calendar, error)
	GetUserDefaultCalendarView(userID string, startTime, endTime time.Time) ([]*Event, error)
}

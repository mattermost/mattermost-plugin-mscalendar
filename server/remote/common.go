// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import "time"

type EmailAddress struct {
	Address string `json:"address"`
	Name    string `json:"name,omitempty"`
}
type DateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone,omitempty"`
}

const RFC3339NanoNoTimezone = "2006-01-02T15:04:05.999999999"

func NewDateTime(t time.Time) *DateTime {
	return &DateTime{
		DateTime: t.Format(RFC3339NanoNoTimezone),
		// TimeZone: t.Format("MST"),
		TimeZone: "Eastern Standard Time",

	}
}

func (dt DateTime) String() string {
	t := dt.Time()
	if t.IsZero() {
		return "n/a"
	}
	return t.Format(time.RFC3339)
}

func (dt DateTime) Time() time.Time {
	loc, err := time.LoadLocation(dt.TimeZone)
	if err != nil {
		return time.Time{}
	}

	t, err := time.ParseInLocation(RFC3339NanoNoTimezone, dt.DateTime, loc)
	if err != nil {
		return time.Time{}
	}
	return t
}

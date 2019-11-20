// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import "time"

type DateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone,omitempty"`
}

func NewDateTime(t time.Time) *DateTime {
	return &DateTime{
		DateTime: t.Format("2006-01-02T15:04:05.999999999"), // time.RFC3339Nano sans timezone
		TimeZone: t.Format("MST"),
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
		panic(err.Error())
		return time.Time{}
	}

	t, err := time.ParseInLocation("2006-01-02T15:04:05.999999999", dt.DateTime, loc)
	if err != nil {
		panic(err.Error())
		return time.Time{}
	}
	return t
}

type EmailAddress struct {
	Address string `json:"address"`
	Name    string `json:"name,omitempty"`
}

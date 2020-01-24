// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

type User struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName,omitempty"`
	UserPrincipalName string `json:"userPrincipalName,omitempty"`
	Mail              string `json:"mail,omitempty"`
}

type WorkingHours struct {
	DaysOfWeek []string `json:"daysOfWeek"`
	StartTime  string   `json:"startTime"`
	EndTime    string   `json:"endTime"`
	TimeZone   struct {
		Name string `json:"name"`
	}
}

type MailboxSettings struct {
	TimeZone     string       `json:"timeZone"`
	WorkingHours WorkingHours `json:"workingHours"`
}

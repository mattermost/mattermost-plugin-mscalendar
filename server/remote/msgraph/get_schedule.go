// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"io/ioutil"
	"encoding/json"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"

	msgraph "github.com/yaegashi/msgraph.go/v1.0"
)

type GetScheduleRequest struct {
	// List of emails of users that we want to check
	ScheduleIDs              []string        `json:"schedules"`

	// Overall start and end of entire search window
	StartTime                remote.DateTime `json:"startTime"`
	EndTime                  remote.DateTime `json:"endTime"`

	// Size of each chunk of time we want to check
	// This can be equal to end - start if we want, or we can get more granular results by making it shorter.
	// For the graph API: The default is 30 minutes, minimum is 6, maximum is 1440
	// 15 is currently being used on our end
	AvailabilityViewInterval int             `json:"availabilityViewInterval"`
}

func (c *client) GetSchedule(scheduleIDs []string, startTime, endTime *remote.DateTime, availabilityViewInterval int) ([]*remote.ScheduleInformation, error) {
	req := &msgraph.CalendarGetScheduleRequestParameter{
		Schedules: scheduleIDs,
		StartTime: &msgraph.DateTimeTimeZone{
			DateTime: &startTime.DateTime,
			TimeZone: &startTime.TimeZone,
		},
		EndTime: &msgraph.DateTimeTimeZone{
			DateTime: &endTime.DateTime,
			TimeZone: &endTime.TimeZone,
		},
		AvailabilityViewInterval: &availabilityViewInterval,
	}

	// req := &GetScheduleRequest{
	// 	Schedules: scheduleIDs,
	// 	StartTime: startTime,
	// 	EndTime: endTime,
	// 	AvailabilityViewInterval: availabilityViewInterval,
	// }

	r2 := c.rbuilder.Me().Calendar().GetSchedule(req).Request()
	r, err := r2.NewJSONRequest("POST", "", req)
	if err != nil {
		return nil, err
	}
	r = r.WithContext(c.ctx)
	res, err := r2.Client().Do(r)

	if err != nil {
		return nil, err
	}

	var resBody remote.GetScheduleResponse
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &resBody)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return resBody.Value, nil
}

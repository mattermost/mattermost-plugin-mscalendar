package msgraph

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

func (c *appClient) GetSchedule(schedules []string, startTime, endTime *remote.DateTime, availabilityViewInterval int) ([]*remote.ScheduleInformation, error) {
	token, err := c.getAppLevelToken()
	if err != nil {
		return nil, err
	}

	var res remote.GetScheduleResponse

	params := &GetScheduleRequest{
		Schedules: schedules,
		StartTime: startTime,
		EndTime: endTime,
		AvailabilityViewInterval: availabilityViewInterval,
	}

	uid := "fb10ac13-e441-4611-8431-8ee3b6403673"
	u := "https://graph.microsoft.com/v1.0/Users/" + uid + "/calendar/getSchedule"
	_, err = c.Call(http.MethodPost, u, token, params, &res)
	if err != nil {
		return nil, err
	}

	return res.Value, nil
}
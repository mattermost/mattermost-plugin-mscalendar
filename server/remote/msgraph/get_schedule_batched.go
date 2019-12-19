package msgraph

import (
	"strconv"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

type GetScheduleSingleResponse struct {
	ID      string                      `json:"id"`
	Status  int                         `json:"status"`
	Body    *remote.GetScheduleResponse `json:"body"`
	Headers map[string]string           `json:"headers"`
}

type GetScheduleBatchResponse struct {
	Responses []*GetScheduleSingleResponse `json:"responses"`
}

type GetScheduleRequest struct {
	// List of emails of users that we want to check
	Schedules []string `json:"schedules"`

	// Overall start and end of entire search window
	StartTime *remote.DateTime `json:"startTime"`
	EndTime   *remote.DateTime `json:"endTime"`

	// Size of each chunk of time we want to check
	// This can be equal to end - start if we want, or we can get more granular results by making it shorter.
	// For the graph API: The default is 30 minutes, minimum is 6, maximum is 1440
	// 15 is currently being used on our end
	AvailabilityViewInterval int `json:"availabilityViewInterval"`
}

func (c *client) GetSchedule(remoteUserID string, schedules []string, startTime, endTime *remote.DateTime, availabilityViewInterval int) ([]*remote.ScheduleInformation, error) {
	params := &GetScheduleRequest{
		StartTime:                startTime,
		EndTime:                  endTime,
		AvailabilityViewInterval: availabilityViewInterval,
	}

	allRequests := getFullBatchRequest(remoteUserID, schedules, params)

	batchRes := GetScheduleBatchResponse{}
	err := c.batchRequest(allRequests, &batchRes)
	if err != nil {
		return nil, err
	}

	sorted := make([]*GetScheduleSingleResponse, len(allRequests))
	for _, r := range batchRes.Responses {
		id, _ := strconv.Atoi(r.ID)
		sorted[id] = r
	}

	result := []*remote.ScheduleInformation{}
	for _, r := range sorted {
		for _, sched := range r.Body.Value {
			result = append(result, sched)
		}
	}

	return result, nil
}

func getFullBatchRequest(remoteUserID string, schedules []string, params *GetScheduleRequest) []*SingleRequest {
	u := "/Users/" + remoteUserID + "/calendar/getSchedule"

	makeRequest := func() *SingleRequest {
		p := &GetScheduleRequest{
			Schedules:                schedules, // need to chunk these out
			StartTime:                params.StartTime,
			EndTime:                  params.EndTime,
			AvailabilityViewInterval: params.AvailabilityViewInterval,
		}
		req := &SingleRequest{
			URL:    u,
			Method: "POST",
			Body:   p,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
		return req
	}

	//  This is where we can simulate large batches
	// 	TODO: Split up emails given into different batches properly
	allRequests := []*SingleRequest{}
	// allRequests = append(allRequests, makeRequest())

	numRequestsInBatch := 1
	// numRequestsInBatch := 20

	for i := 0; i < numRequestsInBatch; i++ {
		allRequests = append(allRequests, makeRequest())
	}

	for i, r := range allRequests {
		r.ID = strconv.Itoa(i)
	}

	return allRequests
}

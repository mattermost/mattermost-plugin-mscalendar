package msgraph

import (
	"strconv"

	"github.com/mitchellh/mapstructure"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

const maxNumUsersPerRequest = 20

type getScheduleSingleResponse struct {
	ID      string              `json:"id"`
	Status  int                 `json:"status"`
	Body    getScheduleResponse `json:"body"`
	Headers map[string]string   `json:"headers"`
}

type getScheduleBatchResponse struct {
	Responses []*getScheduleSingleResponse `json:"responses"`
}

type getScheduleRequest struct {
	// List of emails of users that we want to check
	Schedules []string `json:"schedules"`

	// Overall start and end of entire search window
	StartTime *remote.DateTime `json:"startTime"`
	EndTime   *remote.DateTime `json:"endTime"`

	/*
		Size of each chunk of time we want to check
		This can be equal to end - start if we want, or we can get more granular results by making it shorter.
		For the graph API: The default is 30 minutes, minimum is 6, maximum is 1440
		15 is currently being used on our end
	*/
	AvailabilityViewInterval int `json:"availabilityViewInterval"`
}

type getScheduleResponse struct {
	Value []*remote.ScheduleInformation `json:"value,omitempty"`
}

func (c *client) GetSchedule(remoteUserID string, schedules []string, startTime, endTime *remote.DateTime, availabilityViewInterval int) ([]*remote.ScheduleInformation, error) {
	params := &getScheduleRequest{
		StartTime:                startTime,
		EndTime:                  endTime,
		AvailabilityViewInterval: availabilityViewInterval,
	}

	allRequests := prepareGetScheduleRequests(remoteUserID, schedules, params)

	err, batchResponses := c.batchRequest(allRequests)
	if err != nil {
		return nil, err
	}

	result := []*remote.ScheduleInformation{}

	for i, batchRes := range batchResponses {
		length := maxNumRequestsPerBatch
		if i == len(batchResponses)-1 {
			length = len(allRequests) % maxNumRequestsPerBatch
		}

		sorted := make([]*getScheduleSingleResponse, length)

		for _, r := range batchRes.Responses {
			res := &getScheduleSingleResponse{}
			mapstructure.Decode(r, res)

			id, _ := strconv.Atoi(res.ID)
			sorted[id] = res
		}

		for _, r := range sorted {
			for _, sched := range r.Body.Value {
				result = append(result, sched)
			}
		}
	}

	return result, nil
}

func prepareGetScheduleRequests(remoteUserID string, schedules []string, params *getScheduleRequest) []*singleRequest {
	u := "/Users/" + remoteUserID + "/calendar/getSchedule"

	makeRequest := func(schedBatch []string) *singleRequest {
		req := &singleRequest{
			URL:    u,
			Method: "POST",
			Body: &getScheduleRequest{
				Schedules:                schedBatch,
				StartTime:                params.StartTime,
				EndTime:                  params.EndTime,
				AvailabilityViewInterval: params.AvailabilityViewInterval,
			},
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
		return req
	}

	allRequests := []*singleRequest{}

	numUsers := len(schedules)
	numRequests := numUsers / maxNumUsersPerRequest
	if numUsers%maxNumUsersPerRequest != 0 {
		numRequests += 1
	}

	for i := 0; i < numRequests; i++ {
		startIdx := i * maxNumUsersPerRequest
		endIdx := startIdx + maxNumUsersPerRequest
		if i == numRequests-1 {
			endIdx = len(schedules)
		}

		slice := schedules[startIdx:endIdx]
		req := makeRequest(slice)
		req.ID = strconv.Itoa(i)
		allRequests = append(allRequests, req)
	}

	return allRequests
}

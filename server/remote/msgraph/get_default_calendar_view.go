// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

type calendarViewResponse struct {
	Error *remote.APIError `json:"error,omitempty"`
	Value []*remote.Event  `json:"value,omitempty"`
}

type calendarViewSingleResponse struct {
	Headers map[string]string    `json:"headers"`
	ID      string               `json:"id"`
	Body    calendarViewResponse `json:"body"`
	Status  int                  `json:"status"`
}

type calendarViewBatchResponse struct {
	Responses []*calendarViewSingleResponse `json:"responses"`
}

func (c *client) GetDefaultCalendarView(remoteUserID string, start, end time.Time) ([]*remote.Event, error) {
	paramStr := getQueryParamStringForCalendarView(start, end)

	res := &calendarViewResponse{}
	err := c.rbuilder.Users().ID(remoteUserID).CalendarView().Request().JSONRequest(
		c.ctx, http.MethodGet, paramStr, nil, res)
	if err != nil {
		return nil, errors.Wrap(err, "msgraph GetDefaultCalendarView")
	}

	return res.Value, nil
}

func (c *client) DoBatchViewCalendarRequests(allParams []*remote.ViewCalendarParams) ([]*remote.ViewCalendarResponse, error) {
	requests := []*singleRequest{}
	for _, params := range allParams {
		u := getCalendarViewURL(params)
		req := &singleRequest{
			ID:      params.RemoteUserID,
			URL:     u,
			Method:  http.MethodGet,
			Headers: map[string]string{},
		}
		requests = append(requests, req)
	}

	batchRequests := prepareBatchRequests(requests)
	var batchResponses []*calendarViewBatchResponse
	for _, req := range batchRequests {
		batchRes := &calendarViewBatchResponse{}
		err := c.batchRequest(req, batchRes)
		if err != nil {
			return nil, errors.Wrap(err, "msgraph ViewCalendar batch request")
		}

		batchResponses = append(batchResponses, batchRes)
	}

	result := []*remote.ViewCalendarResponse{}
	for _, batchRes := range batchResponses {
		for _, res := range batchRes.Responses {
			viewCalRes := &remote.ViewCalendarResponse{
				RemoteUserID: res.ID,
				Events:       res.Body.Value,
				Error:        res.Body.Error,
			}
			result = append(result, viewCalRes)
		}
	}

	return result, nil
}

func getCalendarViewURL(params *remote.ViewCalendarParams) string {
	paramStr := getQueryParamStringForCalendarView(params.StartTime, params.EndTime)
	return "/Users/" + params.RemoteUserID + "/calendarView" + paramStr
}

func getQueryParamStringForCalendarView(start, end time.Time) string {
	q := url.Values{}
	q.Add("startDateTime", start.Format(time.RFC3339))
	q.Add("endDateTime", end.Format(time.RFC3339))
	q.Add("$top", "20")
	return "?" + q.Encode()
}

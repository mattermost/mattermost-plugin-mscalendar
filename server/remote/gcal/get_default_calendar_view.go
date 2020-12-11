// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package gcal

import (
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

const (
	RemoteEventBusy = "busy"
	RemoteEventFree = "free"

	GoogleEventBusy = "opaque"
	GoogleEventFree = "transparent"

	ResponseYes   = "accepted"
	ResponseMaybe = "tentativelyAccepted"
	ResponseNo    = "declined"
	ResponseNone  = "notResponded"

	GoogleResponseStatusYes   = "accepted"
	GoogleResponseStatusMaybe = "tentative"
	GoogleResponseStatusNo    = "declined"
	GoogleResponseStatusNone  = "needsAction"
)

var responseStatusConversion = map[string]string{
	GoogleResponseStatusYes:   ResponseYes,
	GoogleResponseStatusMaybe: ResponseMaybe,
	GoogleResponseStatusNo:    ResponseNo,
	GoogleResponseStatusNone:  ResponseNone,
}

func (c *client) GetDefaultCalendarView(remoteUserID string, start, end time.Time) ([]*remote.Event, error) {
	service, err := calendar.New(c.httpClient)
	if err != nil {
		return nil, errors.Wrap(err, "gcal GetDefaultCalendarView, error creating service")
	}

	req := service.Events.List("primary")
	req.MaxResults(20)
	req.TimeMin(start.Format(time.RFC3339))
	req.TimeMax(end.Format(time.RFC3339))
	req.SingleEvents(true)
	req.OrderBy("startTime")

	events, err := req.Do()
	if err != nil {
		return nil, errors.Wrap(err, "gcal GetDefaultCalendarView, error performing request")
	}

	result := []*remote.Event{}
	if len(events.Items) == 0 {
		return result, nil
	}

	for _, event := range events.Items {
		if event.ICalUID != "" {
			result = append(result, convertGCalEventToRemoteEvent(event))
		}
	}

	return result, nil
}

func convertGCalEventToRemoteEvent(event *calendar.Event) *remote.Event {
	showAs := RemoteEventBusy
	if event.Transparency == GoogleEventFree {
		showAs = RemoteEventFree
	}

	start := remote.NewGoogleDateTime(event.Start)
	end := remote.NewGoogleDateTime(event.End)

	location := &remote.Location{
		DisplayName: event.Location,
	}

	organizer := &remote.Attendee{
		EmailAddress: &remote.EmailAddress{
			Name:    event.Organizer.Email,
			Address: event.Organizer.Email,
		},
	}

	var responseStatus *remote.EventResponseStatus
	responseRequested := false
	isOrganizer := false

	attendees := []*remote.Attendee{}
	for _, attendee := range event.Attendees {
		attendees = append(attendees, &remote.Attendee{
			Status: &remote.EventResponseStatus{
				Response: attendee.ResponseStatus,
			},
			EmailAddress: &remote.EmailAddress{
				Name:    attendee.Email,
				Address: attendee.Email,
			},
		})

		if attendee.Self {
			if attendee.ResponseStatus == GoogleResponseStatusNone {
				responseRequested = true
			}

			response := responseStatusConversion[attendee.ResponseStatus]
			responseStatus = &remote.EventResponseStatus{
				Response: response,
			}

			isOrganizer = attendee.Organizer
		}
	}

	isAllDay := len(event.Start.Date) > 0 // if Date field is present, it is all-day. as opposed to DateTime field

	return &remote.Event{
		ID:                event.Id,
		ICalUID:           event.ICalUID,
		Subject:           event.Summary,
		Body:              &remote.ItemBody{Content: event.Description},
		BodyPreview:       event.Description, // GCAL TODO no body preview available?
		IsAllDay:          isAllDay,
		ShowAs:            showAs,
		Weblink:           event.HtmlLink,
		Start:             start,
		End:               end,
		Location:          location,
		Organizer:         organizer,
		Attendees:         attendees,
		ResponseStatus:    responseStatus,
		IsCancelled:       event.Status == "cancelled",
		IsOrganizer:       isOrganizer,
		ResponseRequested: responseRequested,
		// 	Importance                 string
		// 	ReminderMinutesBeforeStart int
	}
}

/*
	Rest of file is unimplemented batch request stuff
*/

type calendarViewResponse struct {
	Value []*remote.Event  `json:"value,omitempty"`
	Error *remote.APIError `json:"error,omitempty"`
}

type calendarViewSingleResponse struct {
	ID      string               `json:"id"`
	Status  int                  `json:"status"`
	Body    calendarViewResponse `json:"body"`
	Headers map[string]string    `json:"headers"`
}

type calendarViewBatchResponse struct {
	Responses []*calendarViewSingleResponse `json:"responses"`
}

func (c *client) DoBatchViewCalendarRequests(allParams []*remote.ViewCalendarParams) ([]*remote.ViewCalendarResponse, error) {
	if true {
		return nil, errors.New("gcal DoBatchViewCalendarRequests not implemented")
	}

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

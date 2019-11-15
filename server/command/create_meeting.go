// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jkrecek/msgraph-go"
	"github.com/mattermost/mattermost-plugin-msoffice/server/msgraph"
)

func (h *Handler) createMeeting(parameters ...string) (string, error) {

	if len(parameters) != 3 {
		// return "jason", errors.New("Please specify an issue kassignee>`.")
		return "Run `/msoffice` for general help, or `/msoffice viewcal` to get calendar id", nil
	}

	subject := parameters[0]
	body := parameters[1]
	calId := parameters[2]

	user, err := h.loadRemoteUser()
	if err != nil {
		return "", err
	}

	msgraphClient := msgraph.NewClient(h.Config, user.OAuth2Token)

	eventBody := graph.NewGraphBody(body)
	start := graph.NewGraphTime(time.Now())
	end := graph.NewGraphTime(time.Now().Add(time.Hour))

	start.TimeZone = "UTC"
	end.TimeZone = "UTC"

	event := &graph.Event{}
	event.Subject = subject
	event.BodyPreview = "testBodyPreview"
	event.Body = eventBody
	event.Start = start
	event.End = end

	fmt.Printf("event = %+v\n", event)

	calEvent, err := msgraphClient.CreateCalendarEvent(calId, event)
	if err != nil {
		return "", err
	}
	fmt.Printf("calEvent = %+v\n", calEvent)

	bb, _ := json.MarshalIndent(event, "", "  ")
	resp := "<><>" + string(bb)
	return resp, nil
}

func (h *Handler) getEvents(parameters ...string) (string, error) {
	user, err := h.loadRemoteUser()
	if err != nil {
		return "", err
	}

	msgraphClient := msgraph.NewClient(h.Config, user.OAuth2Token)

	id := "AQMkADAwATNiZmYAZC04OTAyLWQ5MjMtMDACLTAwCgBGAAADftmoUtfD8EmVrSDo_TkEIAcAh9lM3vJmd0exkjhGMF083wAAAgEGAAAAh9lM3vJmd0exkjhGMF083wAAAjwNAAAA"
	cals, err := msgraphClient.GetCalendarEvents(id)
	if err != nil {
		return "", err
	}
	bb, _ := json.MarshalIndent(cals, "", "  ")

	resp := "<><>" + string(bb)
	return resp, nil
}

func (h *Handler) createCalendar(parameters ...string) (string, error) {

	if len(parameters) != 1 {
		return "please specify a calendar name", nil
	}
	calName := parameters[0]

	user, err := h.loadRemoteUser()
	if err != nil {
		return "", err
	}

	msgraphClient := msgraph.NewClient(h.Config, user.OAuth2Token)

	newCal := &graph.Calendar{
		Name: calName,
	}

	cal, err := msgraphClient.CreateCalendar(newCal)
	if err != nil {
		return "", err
	}
	bb, _ := json.MarshalIndent(cal, "", "  ")

	resp := "Calendar has been created: " + calName + "\n"
	resp += "" + string(bb)
	return resp, nil
}

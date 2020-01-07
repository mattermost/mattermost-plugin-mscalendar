package command

import (
	"fmt"
	"strings"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
	flag "github.com/spf13/pflag"
)

func getCreateEventFlagSet() *flag.FlagSet {
	flagSet := flag.NewFlagSet("create", flag.ContinueOnError)
	flagSet.String("subject", "", "Subject of the Event (no spaces for now)")
	flagSet.String("starttime", time.Now().Format(time.RFC3339), "Start time for the event")
	flagSet.String("endtime", time.Now().Add(time.Hour).Format(time.RFC3339), "End time for the event")
	flagSet.StringSlice("attendees", []string{}, "A comma separated list of Attendees")

	return flagSet
}

type userError struct {
	ErrorMessage string
}

func parseCreateArgs(args []string) (*remote.Event, *userError, error) {

	var attendees []*remote.Attendee

	attendee1 := &remote.Attendee{
		Type: "required",
		Status: &remote.EventResponseStatus{
			Response: "",
			Time:     "",
		},
		EmailAddress: &remote.EmailAddress{
			Address: "joe@example.com",
			Name:    "joe smith",
		},
	}

	attendee2 := &remote.Attendee{
		Type: "required",
		Status: &remote.EventResponseStatus{
			Response: "",
			Time:     "",
		},
		EmailAddress: &remote.EmailAddress{
			Address: "jane@example.com",
			Name:    "jane smith",
		},
	}

	attendees = append(attendees, attendee1)
	attendees = append(attendees, attendee2)

	// create event
	event := &remote.Event{
		Subject: "TestSubject",
		// BodyPreview: "testBodyPreview",
		BodyPreview: "DEBUG_BodyPreview",
		Body: &remote.ItemBody{
			Content:     "Hello!  Here is the start of Your Body!",
			ContentType: "Text",
		},
		ReminderMinutesBeforeStart: 15,
		Location: &remote.Location{
			DisplayName:  "Las Vegas",
			LocationType: "homeAddress",
			Address: &remote.Address{
				Street:          "3730 Las Vegas Blvd S",
				City:            "Las Vegas",
				State:           "Nevada",
				CountryOrRegion: "US",
				PostalCode:      "89158",
			},
			Coordinates: &remote.Coordinates{
				Latitude:  47.672,
				Longitude: -102.103,
			},
		},
		Attendees: attendees,
		Start: &remote.DateTime{
			TimeZone: "Pacific Standard Time",
			DateTime: time.Now().Format(time.RFC3339),
		},
		End: &remote.DateTime{
			TimeZone: "Pacific Standard Time",
			DateTime: time.Now().Add(time.Hour).Format(time.RFC3339),
		},
	}

	// parse flags and start overriding Demo Defaults
	createFlagSet := getCreateEventFlagSet()
	err := createFlagSet.Parse(args)
	if err != nil {
		return event, nil, err
	}

	//                //
	// Required Flags //
	//                //
	subject, err := createFlagSet.GetString("subject")
	if err != nil {
		return event, nil, err
	}
	// check that next arg is not a flag "--"
	if subject == "" || strings.HasPrefix(subject, "--") {
		return event, &userError{ErrorMessage: "must specify an event subject"}, nil
	}
	event.Subject = subject

	//                //
	// Optional Flags //
	//                //
	startTime, err := createFlagSet.GetString("starttime")
	if err != nil {
		return event, nil, err
	}
	if strings.HasPrefix(startTime, "--") {
		return event, &userError{ErrorMessage: "must specify an event subject"}, nil
	}
	event.Start.DateTime = startTime

	endTime, err := createFlagSet.GetString("endtime")
	if err != nil {
		return event, nil, err
	}
	if strings.HasPrefix(endTime, "--") {
		return event, &userError{ErrorMessage: "must specify an event subject"}, nil
	}
	event.End.DateTime = endTime
	// for a := range event.myattendees {
	// 	fmt.Printf("myattendees[a] = %+v\n", myattendees[a])
	// }
	// attendees, err := createFlagSet.GetStringSlice("attendees")
	// if err != nil {
	// 	return "", nil, nil, err
	// }
	// if len(attendees) == 0 {
	// 	return "must specify some attendees ", nil, nil, nil
	// }

	return event, nil, nil
}

func (c *Command) createEvent(parameters ...string) (string, error) {

	if len(parameters) == 0 {
		return fmt.Sprintf(getCreateEventFlagSet().FlagUsages()), nil
	}

	event, userError, err := parseCreateArgs(parameters)
	if err != nil {
		return "", err
	}
	if userError != nil {
		return string(userError.ErrorMessage), nil
	}

	calEvent, err := c.API.CreateEvent(event)
	if err != nil {
		return "", err
	}
	resp := "Event Created\n" + utils.JSONBlock(&calEvent)

	return resp, nil
}

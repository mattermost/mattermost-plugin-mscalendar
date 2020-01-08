package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
	flag "github.com/spf13/pflag"
)

func getCreateEventFlagSet() *flag.FlagSet {
	flagSet := flag.NewFlagSet("create", flag.ContinueOnError)
	flagSet.Bool("help", false, "show help")
	flagSet.String("test-subject", "", "Subject of the event (no spaces for now)")
	flagSet.String("test-body", "", "Body of the event (no spaces for now)")
	flagSet.StringSlice("test-location", []string{}, "Location of the event <displayName,street,city,state,postalcode,country> (comma separated; no spaces)")
	flagSet.String("starttime", time.Now().Format(time.RFC3339), "Start time for the event")
	flagSet.Bool("allday", false, "Set as all day event (starttime/endtime must be set to midnight on different days - 2019-12-19T00:00:00-00:00)")
	flagSet.Int("reminder", 15, "Reminder (in minutes)")
	flagSet.String("endtime", time.Now().Add(time.Hour).Format(time.RFC3339), "End time for the event")
	flagSet.StringSlice("attendees", []string{}, "A comma separated list of Mattermost UserIDs")

	return flagSet
}

type userError struct {
	ErrorMessage string
}

func parseCreateArgs(args []string) (*remote.Event, *userError, error) {

	event := &remote.Event{}

	createFlagSet := getCreateEventFlagSet()
	err := createFlagSet.Parse(args)
	if err != nil {
		return event, nil, err
	}

	help, err := createFlagSet.GetBool("help")
	if help == true {
		return nil, &userError{ErrorMessage: fmt.Sprintf(getCreateEventFlagSet().FlagUsages())}, nil
	}

	subject, err := createFlagSet.GetString("test-subject")
	if err != nil {
		return event, nil, err
	}
	// check that next arg is not a flag "--"
	if strings.HasPrefix(subject, "--") {
		return event, &userError{ErrorMessage: "must specify an event subject"}, nil
	}
	event.Subject = subject

	body, err := createFlagSet.GetString("test-body")
	if err != nil {
		return event, nil, err
	}
	// check that next arg is not a flag "--"
	if strings.HasPrefix(body, "--") {
		return event, &userError{ErrorMessage: "must specify an event body"}, nil
	}
	event.Body = &remote.ItemBody{
		Content: body,
	}

	startTime, err := createFlagSet.GetString("starttime")
	if err != nil {
		return event, nil, err
	}
	if strings.HasPrefix(startTime, "--") {
		return event, &userError{ErrorMessage: "must specify an starttime"}, nil
	}
	event.Start = &remote.DateTime{
		DateTime: startTime,
		TimeZone: "Pacific Standard Time",
	}

	endTime, err := createFlagSet.GetString("endtime")
	if err != nil {
		return event, nil, err
	}
	if strings.HasPrefix(endTime, "--") {
		return event, &userError{ErrorMessage: "must specify an endtime"}, nil
	}
	event.End = &remote.DateTime{
		DateTime: endTime,
		TimeZone: "Pacific Standard Time",
	}

	allday, err := createFlagSet.GetBool("allday")
	if err != nil {
		return event, nil, err
	}
	event.IsAllDay = allday

	reminder, err := createFlagSet.GetInt("reminder")
	if err != nil {
		return event, nil, err
	}
	if strings.HasPrefix(strconv.Itoa(int(reminder)), "--") {
		return event, &userError{ErrorMessage: "must specify an reminder"}, nil
	}
	event.ReminderMinutesBeforeStart = reminder

	location, err := createFlagSet.GetStringSlice("test-location")
	if err != nil {
		return event, nil, err
	}
	if location != nil {
		if len(location) != 6 {
			return event, &userError{ErrorMessage: "must specify --test-location with 6 parameters, including a comma for empty values"}, nil
		}
		event.Location = &remote.Location{
			LocationType: "default",
			DisplayName:  location[0],
			Address: &remote.Address{
				Street:          location[1],
				City:            location[2],
				State:           location[3],
				PostalCode:      location[4],
				CountryOrRegion: location[5],
			},
		}
	}

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

	createFlagSet := getCreateEventFlagSet()
	err = createFlagSet.Parse(parameters)
	if err != nil {
		return "", err
	}

	mattermostUserIDs, err := createFlagSet.GetStringSlice("attendees")
	if err != nil {
		return "", err
	}

	calEvent, err := c.API.CreateEvent(event, mattermostUserIDs)
	if err != nil {
		return "", err
	}
	resp := "Event Created\n" + utils.JSONBlock(&calEvent)

	return resp, nil
}

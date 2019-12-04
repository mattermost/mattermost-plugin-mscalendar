package command

import (
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
)

func (c *Command) createEvent(parameters ...string) (string, error) {
	// if len(parameters) != 2 {
	// 	// return "jason", errors.New("Please specify an issue kassignee>`.")
	// 	return "Run `/msoffice` for general help or `/msoffice viewcal` to get calendar id", nil
	// }

	// subject := parameters[0]
	// body := parameters[1]
	// calId := parameters[2]

	// parse start time
	start := &remote.DateTime{
		TimeZone: "Pacific Standard Time",
		DateTime: time.Now().Format(time.RFC3339),
	}

	// parse end time
	end := &remote.DateTime{
		TimeZone: "Pacific Standard Time",
		DateTime: time.Now().Add(time.Hour).Format(time.RFC3339),
	}

	var attendees []*remote.EventAttendee

	attendee1 := &remote.EventAttendee{
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

	attendee2 := &remote.EventAttendee{
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
		Subject: "testsubject",
		// BodyPreview: "testBodyPreview",
		BodyPreview: "DEBUG_BodyPreview",
		Body: &remote.ItemBody{
			Content:     "Hello!  Here is the start of Your Body!",
			ContentType: "Text",
		},
		ReminderMinutesBeforeStart: 15,
		Location: &remote.EventLocation{
			DisplayName:  "Home",
			LocationType: "homeAddress",
			Address: &remote.EventAddress{
				Street:          "E Main St",
				City:            "Redmond",
				State:           "WA",
				CountryOrRegion: "US",
				PostalCode:      "32008",
			},
			Coordinates: &remote.EventCoordinates{
				Latitude:  47.672,
				Longitude: -102.103,
			},
		},
		Attendees: attendees,
		Start:     start,
		End:       end,
	}

	calEvent, err := c.API.CreateEvent(event)
	if err != nil {
		return "", err
	}
	resp := "Event Created\n" + utils.JSONBlock(&calEvent)

	return resp, nil
}

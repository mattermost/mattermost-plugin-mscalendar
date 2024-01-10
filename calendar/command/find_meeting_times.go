// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils"
)

func (c *Command) findMeetings(parameters ...string) (string, bool, error) {
	meetingParams := &remote.FindMeetingTimesParameters{}

	var attendees []remote.Attendee
	for _, parameter := range parameters {
		s := strings.Split(parameter, ":")
		if len(s) != 2 {
			return "", false, fmt.Errorf("error in parameter %s", parameter)
		}
		t, email := s[0], s[1]
		// REVIEW: very small struct being used to fetch meeting times. FindMeetingTimesParameters is a large struct, but only attendees being filled here
		attendee := remote.Attendee{
			Type: t,
			EmailAddress: &remote.EmailAddress{
				Address: email,
			},
		}
		attendees = append(attendees, attendee)
	}
	meetingParams.Attendees = attendees

	meetings, err := c.Engine.FindMeetingTimes(c.user(), meetingParams)
	if err != nil {
		return "", false, err
	}

	timeZone, _ := c.Engine.GetTimezone(c.user())
	resp := ""
	for _, m := range meetings.MeetingTimeSuggestions {
		if timeZone != "" {
			m.MeetingTimeSlot.Start = m.MeetingTimeSlot.Start.In(timeZone)
			m.MeetingTimeSlot.End = m.MeetingTimeSlot.End.In(timeZone)
		}
		resp += utils.JSONBlock(renderMeetingTime(m))
	}

	return resp, false, nil
}

func renderMeetingTime(m *remote.MeetingTimeSuggestion) string {
	start := m.MeetingTimeSlot.Start.PrettyString()
	end := m.MeetingTimeSlot.End.PrettyString()
	return fmt.Sprintf("%s - %s (%s)", start, end, m.MeetingTimeSlot.Start.TimeZone)
}

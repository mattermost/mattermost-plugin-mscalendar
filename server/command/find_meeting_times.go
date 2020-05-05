// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
)

func (c *Command) findMeetings(parameters ...string) (string, bool, error) {
	meetingParams := &remote.FindMeetingTimesParameters{}

	var attendees []remote.Attendee
	for a := range parameters {
		s := strings.Split(parameters[a], ":")
		t, email := s[0], s[1]
		attendee := remote.Attendee{
			Type: t,
			EmailAddress: &remote.EmailAddress{
				Address: email,
			},
		}
		attendees = append(attendees, attendee)
	}
	meetingParams.Attendees = attendees

	meetings, err := c.MSCalendar.FindMeetingTimes(c.user(), meetingParams)
	if err != nil {
		return "", false, err
	}

	timeZone, _ := c.MSCalendar.GetTimezone(c.user())
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

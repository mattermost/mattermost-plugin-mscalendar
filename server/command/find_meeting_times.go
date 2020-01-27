// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"strings"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
)

func (c *Command) findMeetings(parameters ...string) (string, error) {
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

	meetings, err := c.MSCalendar.FindMeetingTimes(meetingParams)
	if err != nil {
		return "", err
	}

	resp := ""
	for _, m := range meetings.MeetingTimeSuggestions {
		resp += utils.JSONBlock(m)
	}

	return resp, nil
}

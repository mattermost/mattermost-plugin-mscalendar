// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"encoding/json"
	"fmt"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
)

func (c *Command) findMeetings(parameters ...string) (string, error) {
	meetingParams := &remote.FindMeetingTimesParameters{}
	// meetings, err := c.API.FindMeetingTimes(time.Now(), time.Now().Add(14*24*time.Hour))
	meetings, err := c.API.FindMeetingTimes(meetingParams)
	if err != nil {
		return "", err
	}

	fmt.Printf("meetings = %+v\n", meetings)
	meetingsD, _ := json.MarshalIndent(meetings, "", "    ")
	fmt.Printf("meetings = %+v\n", string(meetingsD))

	resp := ""
	for _, m := range meetings.MeetingTimeSuggestions {
		resp += "  - " + utils.JSONBlock(m)
	}

	return resp, nil
}

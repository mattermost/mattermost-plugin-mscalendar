// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

// FindMeetingTimes finds Meetinsa calendar event
func (c *client) FindMeetingTimes(findMeetingsParam *remote.FindMeetingTimesParameters) (*remote.MeetingTimeSuggestionResults, error) {
	meetingsOut := &remote.MeetingTimeSuggestionResults{}

	req := c.rbuilder.Me().FindMeetingTimes(nil).Request()
	err := req.JSONRequest(c.ctx, http.MethodPost, "", &findMeetingsParam, &meetingsOut)

	// fakeResD, _ := json.MarshalIndent(fakeRes, "", "    ")
	// fmt.Printf("fakeRes = %+v\n", string(fakeResD))

	if err != nil {
		return nil, err
	}
	return meetingsOut, nil
}

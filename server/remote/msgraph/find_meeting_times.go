// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	msgraph "github.com/yaegashi/msgraph.go/v1.0"
)

// FindMeetingTimes finds Meetinsa calendar event
func (c *client) FindMeetingTimes(findMeetingsParam *remote.FindMeetingTimesParameters) (*remote.MeetingTimeSuggestionResults, error) {
	meetingsOut := &remote.MeetingTimeSuggestionResults{}
	fakeReq := &msgraph.UserFindMeetingTimesRequestParameter{}
	fakeRes := &msgraph.MeetingTimeSuggestionsResult{}

	req := c.rbuilder.Me().FindMeetingTimes(fakeReq).Request()
	err := req.JSONRequest(c.ctx, http.MethodPost, "", &findMeetingsParam, &fakeRes)

	fakeReqD, _ := json.MarshalIndent(fakeReq, "", "    ")
	fmt.Printf("fakeReq = %+v\n", string(fakeReqD))

	fakeResD, _ := json.MarshalIndent(fakeRes, "", "    ")
	fmt.Printf("fakeRes = %+v\n", string(fakeResD))

	if err != nil {
		return nil, err
	}
	return meetingsOut, nil
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

// FindMeetingTimes finds meeting time suggestions for a calendar event
func (c *client) FindMeetingTimes(userID string, params *remote.FindMeetingTimesParameters) (*remote.MeetingTimeSuggestionResults, error) {
	meetingsOut := &remote.MeetingTimeSuggestionResults{}
	req := c.rbuilder.Users().ID(userID).FindMeetingTimes(nil).Request()
	err := req.JSONRequest(c.ctx, http.MethodPost, "", &params, &meetingsOut)
	if err != nil {
		return nil, err
	}
	return meetingsOut, nil
}

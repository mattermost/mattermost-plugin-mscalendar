// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

// CreateCalendar creates a calendar
func (c *client) CreateCalendar(calendarName string) (*remote.Calendar, error) {
	var calendar = &remote.Calendar{
		Name: calendarName,
	}
	var calOut = remote.Calendar{}
	req := c.rbuilder.Me().Calendars().Request()
	err := req.JSONRequest(c.ctx, http.MethodPost, "", &calendar, &calOut)

	calD, _ := json.MarshalIndent(calOut, "", "    ")
	fmt.Printf("cal = %+v\n", string(calD))

	if err != nil {
		return nil, err
	}

	return &calOut, nil
}

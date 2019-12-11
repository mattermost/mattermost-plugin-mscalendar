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
func (c *client) CreateCalendar(calIn *remote.Calendar) (*remote.Calendar, error) {
	var calOut = remote.Calendar{}
	err := c.rbuilder.Me().Calendars().Request().JSONRequest(c.ctx, http.MethodPost, "", &calIn, &calOut)
	if err != nil {
		return nil, err
	}

	calD, _ := json.MarshalIndent(calOut, "", "    ")
	fmt.Printf("cal = %+v\n", string(calD))

	return &calOut, nil
}

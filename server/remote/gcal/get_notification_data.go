// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package gcal

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

func (c *client) GetNotificationData(orig *remote.Notification) (*remote.Notification, error) {
	service, err := calendar.NewService(context.Background(), option.WithHTTPClient(c.httpClient))
	if err != nil {
		return nil, errors.Wrap(err, "gcal GetNotificationData, error creating service")
	}

	n := *orig
	wh := n.Webhook.(*webhook)

	cal, err := c.GetDefaultCalendar()
	if err != nil {
		return nil, errors.Wrap(err, "gcal GetNotificationData, error getting default calendar")
	}
	d, _ := json.Marshal(wh)
	c.Logger.Debugf("%v", string(d))

	result, err := service.Events.List(cal.ID).SyncToken(wh.Resource).MaxResults(10).Do()
	if err != nil {
		return nil, errors.Wrap(err, "gcal GetNotificationData, error getting changed events")
	}

	d2, _ := json.Marshal(result)
	c.Logger.Debugf("events: %v", string(d2))

	reqBody := service.Events.Get(cal.ID, wh.Resource)
	googleEvent, err := reqBody.Do()
	if err != nil {
		return nil, errors.Wrap(err, "gcal GetNotificationData, error fetching event data")
	}

	event := convertGCalEventToRemoteEvent(googleEvent)

	n.Event = event
	n.IsBare = false

	return &n, nil
}

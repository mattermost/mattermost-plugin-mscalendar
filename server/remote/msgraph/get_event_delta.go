// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"net/http"
	"net/url"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

type getEventDeltaResponse struct {
	NextLink  string          `json:"@odata.nextLink"`
	DeltaLink string          `json:"@odata.deltaLink"`
	Value     []*remote.Event `json:"value"`
}

func (c *client) GetEventDeltaFromDateRange(remoteUserID string, start, end *remote.DateTime) (events []*remote.Event, deltaLink string, err error) {
	q := url.Values{}
	q.Add("StartDateTime", start.Time().Format(time.RFC3339))
	q.Add("EndDateTime", end.Time().Format(time.RFC3339))
	params := "?" + q.Encode()

	u := c.rbuilder.Me().CalendarView().URL() + "/delta" + params
	var out getEventDeltaResponse

	_, err = c.CallJSON(http.MethodGet, u, nil, &out)
	if err != nil {
		return nil, "", err
	}

	ls := []*remote.Event{}
	ls = append(ls, out.Value...)

	nextLink := out.NextLink
	for nextLink != "" {
		out = getEventDeltaResponse{}

		_, err = c.CallJSON(http.MethodGet, nextLink, nil, &out)
		if err != nil {
			return nil, "", err
		}

		ls = append(ls, out.Value...)
		nextLink = out.NextLink
	}

	return ls, out.DeltaLink, nil
}

func (c *client) GetEventDeltaFromURL(deltaURL string) (events []*remote.Event, deltaLink string, err error) {
	var out getEventDeltaResponse

	_, err = c.CallJSON(http.MethodGet, deltaURL, nil, &out)
	if err != nil {
		return nil, "", err
	}

	ls := []*remote.Event{}
	ls = append(ls, out.Value...)

	nextLink := out.NextLink
	for nextLink != "" {
		out = getEventDeltaResponse{}

		_, err = c.CallJSON(http.MethodGet, nextLink, nil, &out)
		if err != nil {
			return nil, "", err
		}

		ls = append(ls, out.Value...)
		nextLink = out.NextLink
	}

	return ls, out.DeltaLink, nil
}

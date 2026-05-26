// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/httputils"
)

func (api *api) viewEvents(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-Id")
	if mattermostUserID == "" {
		api.Logger.Errorf("viewEvents, unauthorized user")
		httputils.WriteUnauthorizedError(w, fmt.Errorf("unauthorized"))
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		httputils.WriteBadRequestError(w, fmt.Errorf("from and to query parameters are required"))
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		httputils.WriteBadRequestError(w, fmt.Errorf("invalid from parameter: %w", err))
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		httputils.WriteBadRequestError(w, fmt.Errorf("invalid to parameter: %w", err))
		return
	}

	if from.After(to) {
		httputils.WriteBadRequestError(w, fmt.Errorf("from must be before or equal to to"))
		return
	}

	const maxWindow = 62 * 24 * time.Hour
	if to.Sub(from) > maxWindow {
		httputils.WriteBadRequestError(w, fmt.Errorf("date range must not exceed 62 days"))
		return
	}

	eng := engine.New(api.Env, mattermostUserID)
	user := engine.NewUser(mattermostUserID)

	events, err := eng.ViewCalendar(user, from, to)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			api.Logger.With(bot.LogContext{"err": err.Error()}).Errorf("viewEvents, user not found in store")
			httputils.WriteUnauthorizedError(w, fmt.Errorf("unauthorized"))
			return
		}
		api.Logger.With(bot.LogContext{"err": err.Error()}).Errorf("viewEvents, error fetching calendar events")
		httputils.WriteInternalServerError(w, fmt.Errorf("error fetching calendar events"))
		return
	}

	if events == nil {
		events = []*remote.Event{}
	}

	for _, e := range events {
		remote.NormalizeDateTimeToRFC3339(e)
	}

	httputils.WriteJSONResponse(w, events, http.StatusOK)
}

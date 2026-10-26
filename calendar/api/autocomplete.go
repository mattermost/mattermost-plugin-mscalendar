// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/httputils"
)

const (
	autocompleteLimit       = 10
	visibilityBatchSize     = 100
	maxRestrictedCandidates = 200
)

func (api *api) autocompleteConnectedUsers(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-Id")
	_, err := api.Store.LoadUser(mattermostUserID)
	if mattermostUserID == "" || errors.Is(err, store.ErrNotFound) {
		api.Logger.With(bot.LogContext{"err": err.Error()}).Errorf("user unauthorized")
		httputils.WriteUnauthorizedError(w, fmt.Errorf("unauthorized"))
		return
	}

	restrictions, err := api.PluginAPI.GetViewUsersRestrictions(mattermostUserID)
	if err != nil {
		api.Logger.With(bot.LogContext{"err": err.Error()}).Errorf("unable to filter users")
		httputils.WriteInternalServerError(w, err)
		return
	}

	searchLimit := autocompleteLimit
	if restrictions != nil {
		searchLimit = maxRestrictedCandidates
	}

	searchString := r.URL.Query().Get("search")
	results, err := api.Store.SearchInUserIndex(searchString, searchLimit)
	if err != nil {
		api.Logger.With(bot.LogContext{"err": err.Error()}).Errorf("unable to search in user index")
		utils.SlackAttachmentError(w, "unable to search in user index: "+err.Error())
		httputils.WriteInternalServerError(w, err)
		return
	}

	if restrictions != nil {
		results, err = api.filterVisibleUsers(mattermostUserID, restrictions, results, autocompleteLimit)
		if err != nil {
			api.Logger.With(bot.LogContext{"err": err.Error()}).Errorf("unable to filter users")
			httputils.WriteInternalServerError(w, err)
			return
		}
	}

	if err := httputils.WriteJSONResponse(w, results.ToDTO(), http.StatusOK); err != nil {
		api.Logger.With(bot.LogContext{"err": err.Error()}).Errorf("error sending response to user")
		httputils.WriteInternalServerError(w, err)
	}
}

func (api *api) filterVisibleUsers(viewerID string, restrictions *model.ViewUsersRestrictions, users store.UserIndex, limit int) (store.UserIndex, error) {
	result := store.UserIndex{}
	for start := 0; start < len(users) && len(result) < limit; start += visibilityBatchSize {
		batch := users[start:min(start+visibilityBatchSize, len(users))]

		ids := make([]string, 0, len(batch))
		for _, u := range batch {
			ids = append(ids, u.MattermostUserID)
		}

		visibleIDs, err := api.PluginAPI.FilterVisibleUsers(viewerID, restrictions, ids)
		if err != nil {
			return nil, err
		}

		visible := make(map[string]bool, len(visibleIDs))
		for _, id := range visibleIDs {
			visible[id] = true
		}

		for _, u := range batch {
			if len(result) == limit {
				break
			}
			if visible[u.MattermostUserID] {
				result = append(result, u)
			}
		}
	}
	return result, nil
}

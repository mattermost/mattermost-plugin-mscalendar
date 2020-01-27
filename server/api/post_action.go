// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
)

func preprocessAction(w http.ResponseWriter, req *http.Request) (mscalendar.MSCalendar, string, string) {
	request := model.PostActionIntegrationRequestFromJson(req.Body)
	if request == nil {
		// h.LogWarn("failed to decode PostActionIntegrationRequest")
		http.Error(w, "invalid request", http.StatusBadRequest)
		return nil, "", ""
	}
	eventID, ok := request.Context[config.EventIDKey].(string)
	if !ok {
		// h.LogWarn("no event ID in the request")
		http.Error(w, "invalid request", http.StatusBadRequest)
		return nil, "", ""
	}

	option, _ := request.Context["selected_option"].(string)

	return mscalendar.FromContext(req.Context()), eventID, option
}

func postActionAccept(w http.ResponseWriter, req *http.Request) {
	api, eventID, _ := preprocessAction(w, req)
	if eventID == "" {
		return
	}
	err := api.AcceptEvent(eventID)
	if err != nil {
		// h.LogWarn(err.Error())
		http.Error(w, "Failed to accept event: "+err.Error(), http.StatusBadRequest)
		return
	}
}

func postActionDecline(w http.ResponseWriter, req *http.Request) {
	api, eventID, _ := preprocessAction(w, req)
	if eventID == "" {
		return
	}
	err := api.DeclineEvent(eventID)
	if err != nil {
		// h.LogWarn(err.Error())
		http.Error(w, "Failed to decline event: "+err.Error(), http.StatusBadRequest)
		return
	}
}

func postActionTentative(w http.ResponseWriter, req *http.Request) {
	api, eventID, _ := preprocessAction(w, req)
	if eventID == "" {
		return
	}
	err := api.TentativelyAcceptEvent(eventID)
	if err != nil {
		// h.LogWarn(err.Error())
		http.Error(w, "Failed to tentatively accept event: "+err.Error(), http.StatusBadRequest)
		return
	}
}

func postActionRespond(w http.ResponseWriter, req *http.Request) {
	a, eventID, option := preprocessAction(w, req)
	if eventID == "" {
		return
	}
	err := a.RespondToEvent(eventID, option)
	if err != nil {
		// h.LogWarn(err.Error())
		http.Error(w, "Failed to respond to event: "+err.Error(), http.StatusBadRequest)
		return
	}
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
)

func (api *api) preprocessAction(w http.ResponseWriter, req *http.Request) (mscalendar.MSCalendar, *mscalendar.User, string, string, string) {
	mattermostUserID := req.Header.Get("Mattermost-User-ID")

	request := model.PostActionIntegrationRequestFromJson(req.Body)
	if request == nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return nil, nil, "", "", ""
	}

	eventID, ok := request.Context[config.EventIDKey].(string)
	if !ok {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return nil, nil, "", "", ""
	}
	option, _ := request.Context["selected_option"].(string)
	mscal := mscalendar.New(api.Env, mattermostUserID)

	return mscal, mscalendar.NewUser(mattermostUserID), eventID, option, request.PostId
}

func (api *api) postActionAccept(w http.ResponseWriter, req *http.Request) {
	mscalendar, user, eventID, _, _ := api.preprocessAction(w, req)
	if eventID == "" {
		return
	}
	err := mscalendar.AcceptEvent(user, eventID)
	if err != nil {
		api.Logger.Warnf(err.Error())
		http.Error(w, "Failed to accept event: "+err.Error(), http.StatusBadRequest)
		return
	}
}

func (api *api) postActionDecline(w http.ResponseWriter, req *http.Request) {
	mscalendar, user, eventID, _, _ := api.preprocessAction(w, req)
	if eventID == "" {
		return
	}
	err := mscalendar.DeclineEvent(user, eventID)
	if err != nil {
		http.Error(w, "Failed to decline event: "+err.Error(), http.StatusBadRequest)
		return
	}
}

func (api *api) postActionTentative(w http.ResponseWriter, req *http.Request) {
	mscalendar, user, eventID, _, _ := api.preprocessAction(w, req)
	if eventID == "" {
		return
	}
	err := mscalendar.TentativelyAcceptEvent(user, eventID)
	if err != nil {
		http.Error(w, "Failed to tentatively accept event: "+err.Error(), http.StatusBadRequest)
		return
	}
}

func (api *api) postActionRespond(w http.ResponseWriter, req *http.Request) {
	calendar, user, eventID, option, postID := api.preprocessAction(w, req)
	if eventID == "" {
		return
	}
	err := calendar.RespondToEvent(user, eventID, option)
	if err != nil && !strings.HasPrefix(err.Error(), "202") {
		http.Error(w, "Failed to respond to event: "+err.Error(), http.StatusBadRequest)
		return
	}

	var response string
	switch option {
	case mscalendar.OptionYes:
		response = mscalendar.ResponseYes
	case mscalendar.OptionNo:
		response = mscalendar.ResponseNo
	case mscalendar.OptionMaybe:
		response = mscalendar.ResponseMaybe
	default:
		return
	}

	p, err := api.PluginAPI.GetPost(postID)
	if err != nil {
		http.Error(w, "Failed to update the post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sas := p.Attachments()
	if len(sas) == 0 {
		http.Error(w, "Failed to update the post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sa := sas[0]
	sa.Actions = mscalendar.GetPostActionSelect(eventID, response, req.URL.String())
	postResponse := model.PostActionIntegrationResponse{}
	model.ParseSlackAttachment(p, []*model.SlackAttachment{sa})

	postResponse.Update = p

	w.Header().Set("Content-Type", "application/json")
	w.Write(postResponse.ToJson())
}

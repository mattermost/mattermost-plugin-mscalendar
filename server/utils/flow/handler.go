// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package flow

import (
	"net/http"
	"strconv"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
)

type fh struct {
	flow  Flow
	store FlowStore
}

func Init(h *httputils.Handler, flow Flow, store FlowStore) {
	fh := &fh{
		flow:  flow,
		store: store,
	}

	oauth2Router := h.Router.PathPrefix("/").Subrouter()
	oauth2Router.HandleFunc(flow.URL(), fh.handleFlow).Methods("POST")
}

func (fh *fh) handleFlow(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-ID")
	if mattermostUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	stepNumber, err := strconv.Atoi(r.URL.Query().Get("step"))
	if err != nil {
		http.Error(w, "Step is not an int, err="+err.Error(), http.StatusBadRequest)
		return
	}

	step := fh.flow.Step(stepNumber)
	if step == nil {
		http.Error(w, "Not valid step", http.StatusBadRequest)
		return
	}

	property := step.PropertyName()

	valueString := r.URL.Query().Get(property)
	if valueString == "" {
		http.Error(w, "Correct property not set", http.StatusBadRequest)
		return
	}

	value := valueString == "true"

	err = fh.store.SetProperty(mattermostUserID, property, value)
	if err != nil {
		http.Error(w, "There has been a problem setting the property, err="+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.PostActionIntegrationResponse{}
	post := model.Post{}
	model.ParseSlackAttachment(&post, []*model.SlackAttachment{step.ResponseSlackAttachment(value)})

	response.Update = &post

	w.Header().Set("Content-Type", "application/json")
	w.Write(response.ToJson())

	fh.store.RemovePostID(mattermostUserID, property)
	fh.flow.StepDone(mattermostUserID, value)
}

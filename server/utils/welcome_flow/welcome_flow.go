// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package welcome_flow

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
)

type App interface {
	SetUpdateStatus(mattermostUserID string, updateStatus bool) error
	SetGetConfirmation(mattermostUserID string, getConfirmation bool) error
}

type wf struct {
	app App
}

func Init(h *httputils.Handler, app App) {
	wf := &wf{
		app: app,
	}

	oauth2Router := h.Router.PathPrefix("/welcomeBot").Subrouter()
	oauth2Router.HandleFunc("/updateStatus", wf.updateStatus).Methods("POST")
	oauth2Router.HandleFunc("/setConfirmations", wf.updateConfirmations).Methods("POST")
}

func (wf *wf) updateStatus(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-ID")
	if mattermostUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}
	updateStatus := r.URL.Query().Get("update_status") == "true"
	message := ":thumbsup: Got it! We won't update your status in Mattermost."
	if updateStatus {
		message = ":thumbsup: Got it! We'll automatically update your status in Mattermost."
	}

	err := wf.app.SetUpdateStatus(mattermostUserID, updateStatus)
	if err != nil {
		message = "There has been a problem setting your status"
	}

	response := model.PostActionIntegrationResponse{}

	post := model.Post{}
	model.ParseSlackAttachment(&post, []*model.SlackAttachment{getUpdateStatusAttachments(message)})

	response.Update = &post

	w.Header().Set("Content-Type", "application/json")
	w.Write(response.ToJson())
}

func getUpdateStatusAttachments(message string) *model.SlackAttachment {
	return &model.SlackAttachment{
		Title:   "Update Status",
		Text:    message,
		Actions: []*model.PostAction{},
	}
}

func (wf *wf) updateConfirmations(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-ID")
	if mattermostUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}
	getConfirmation := r.URL.Query().Get("get_confirmation") == "true"
	message := "Cool, we will automatically update your status."
	if getConfirmation {
		message = "Cool, we'll also send you confirmations before updating your status."
	}

	err := wf.app.SetGetConfirmation(mattermostUserID, getConfirmation)
	if err != nil {
		message = "There has been a problem setting the confirmation configuration"
	}

	response := model.PostActionIntegrationResponse{}

	post := model.Post{}
	model.ParseSlackAttachment(&post, []*model.SlackAttachment{getSetConfirmationAttachments(message)})

	response.Update = &post

	w.Header().Set("Content-Type", "application/json")
	w.Write(response.ToJson())
}

func getSetConfirmationAttachments(message string) *model.SlackAttachment {
	return &model.SlackAttachment{
		Title:   "Confirm status change",
		Text:    message,
		Actions: []*model.PostAction{},
	}
}

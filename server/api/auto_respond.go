// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
)

func (api *api) autoRespond(w http.ResponseWriter, req *http.Request) {
	mattermostUserID := req.Header.Get("Mattermost-User-ID")
	if mattermostUserID == "" {
		dialogResponseError(w, "Not authorized.")
		return
	}

	v := model.SubmitDialogRequest{}
	err := json.NewDecoder(req.Body).Decode(&v)
	if err != nil {
		dialogResponseError(w, "Failed to process submit dialog response")
		return
	}

	m := mscalendar.New(api.Env, mattermostUserID)
	message, ok := v.Submission["auto_respond"].(string)
	if !ok {
		dialogResponseError(w, `No recognizable value for property "auto_respond".`)
		return
	}

	err = m.SetUserAutoRespondMessage(mattermostUserID, message)
	if err != nil {
		api.Logger.Warnf("Failed to set autorespond message. err=%v", err)
		dialogResponseError(w, "Failed to set autorespond message")
		return
	}

	response := model.SubmitDialogResponse{}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response.ToJson())
	api.Env.Poster.Ephemeral(mattermostUserID, v.ChannelId, "Autorespond message changed to: '%s'", message)
}

func dialogResponseError(w http.ResponseWriter, errorMessage string) {
	response := model.SubmitDialogResponse{
		Error: errorMessage,
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response.ToJson())
}

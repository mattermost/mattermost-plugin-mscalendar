// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package http

import (
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-server/model"

	"github.com/mattermost/mattermost-plugin-msoffice/server/user"
)

func (h *Handler) oauth2Connect(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	conf := h.Remote.NewOAuth2Config(h.Config)
	state := fmt.Sprintf("%v_%v", model.NewId()[0:15], userID)
	stateStore := user.NewOAuth2StateStore(h.API)
	err := stateStore.Store(state)
	if err != nil {
		h.jsonError(w, err)
		return
	}

	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

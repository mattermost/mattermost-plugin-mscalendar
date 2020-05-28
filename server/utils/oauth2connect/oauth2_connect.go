// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package oauth2connect

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
	"github.com/mattermost/mattermost-server/v5/model"
	"golang.org/x/oauth2"
)

func (o *oAuther) oauth2Connect(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	token, _ := o.GetToken(userID)
	if token != nil {
		httputils.WriteInternalServerError(w, errors.New("User already has a token"))
		return
	}

	state := fmt.Sprintf("%v_%v", model.NewId()[0:15], userID)
	appErr := o.api.KVSetWithExpiry(o.getStateKey(userID), []byte(state), oAuth2StateTimeToLive)
	if appErr != nil {
		httputils.WriteInternalServerError(w, errors.New("failed to store token state"))
		return
	}

	redirectURL := o.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

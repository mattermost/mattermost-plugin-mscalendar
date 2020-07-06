// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
)

func (api *api) handleAdminConsent(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-ID")
	if mattermostUserID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Please log into Mattermost."))
		return
	}

	m := mscalendar.New(api.Env, mattermostUserID)
	err := m.VerifyAdminConsentToken(r.URL.Query().Get("state"), mattermostUserID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Failed to verify admin consent token."))
		return
	}

	errStr := r.URL.Query().Get("error")
	if errStr != "" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("The permission grant process has been cancelled."))

		errDesc := r.URL.Query().Get("error_description")
		api.Logger.Warnf("Error received from admin consent redirect. err=%s %s", errStr, errDesc)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Permissions have been granted."))
}

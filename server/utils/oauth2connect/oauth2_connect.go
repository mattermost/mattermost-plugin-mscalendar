// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package oauth2connect

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
	"net/http"
)

func (oa *oa) oauth2Connect(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-ID")
	if mattermostUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	redirectURL, err := oa.app.InitOAuth2(mattermostUserID)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (oa *oa) oauth2ConnectBot(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-ID")
	if mattermostUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	redirectURL, err := oa.app.InitOAuth2ForBot(mattermostUserID)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

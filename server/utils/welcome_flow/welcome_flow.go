// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package welcome_flow

import (
	"fmt"
	"net/http"

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
	oauth2Router.HandleFunc("/updateStatus", wf.updateStatus).Methods("GET")
	oauth2Router.HandleFunc("/setConfirmations", wf.updateConfirmations).Methods("GET")
}

func (wf *wf) updateStatus(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-ID")
	if mattermostUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}
	updateStatus := r.URL.Query().Get("update_status") == "true"
	message := "Status will be updated"
	if !updateStatus {
		message = "Status won't be updated"
	}

	err := wf.app.SetUpdateStatus(mattermostUserID, updateStatus)
	if err != nil {
		message = "There has been a problem setting your status"
	}

	html := `
		<!DOCTYPE html>
		<html>
			<head>
				<script>
					window.close();
				</script>
			</head>
			<body>
				<p>%s. Please close this window.</p>
			</body>
		</html>
		`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(html, message)))
}

func (wf *wf) updateConfirmations(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-ID")
	if mattermostUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}
	getConfirmation := r.URL.Query().Get("get_confirmation") == "true"
	message := "We will ask for confirmation before updating your status"
	if !getConfirmation {
		message = "We will update your status automatically"
	}

	err := wf.app.SetGetConfirmation(mattermostUserID, getConfirmation)
	if err != nil {
		message = "There has been a problem setting the confirmation configuration"
	}

	html := `
		<!DOCTYPE html>
		<html>
			<head>
				<script>
					window.close();
				</script>
			</head>
			<body>
				<p>%s. Please close this window.</p>
			</body>
		</html>
		`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(html, message)))
}

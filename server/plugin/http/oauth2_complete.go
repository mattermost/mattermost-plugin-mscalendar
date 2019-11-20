// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package http

import (
	"github.com/mattermost/mattermost-plugin-msoffice/server/api"
	"net/http"
)

func (h *Handler) oauth2Complete(w http.ResponseWriter, r *http.Request) {
	api := api.FromContext(r.Context())
	mattermostUserID := r.Header.Get("Mattermost-User-ID")
	if mattermostUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	code := r.URL.Query().Get("code")
	if len(code) == 0 {
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")

	err := api.CompleteOAuth2(mattermostUserID, code, state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
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
				<p>Completed connecting to Microsoft Office. Please close this window.</p>
			</body>
		</html>
		`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

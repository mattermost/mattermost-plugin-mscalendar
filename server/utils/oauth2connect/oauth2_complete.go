// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package oauth2connect

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
)

func (oa *oa) oauth2Complete(w http.ResponseWriter, r *http.Request) {
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

	err := oa.app.CompleteOAuth2(mattermostUserID, code, state)
	if err != nil {
		httputils.WriteUnauthorizedError(w, err)
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
				<p>Completed connecting to Microsoft Calendar. Please close this window.</p>
			</body>
		</html>
		`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

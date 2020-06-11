// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package oauther

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
)

func (o *oAuther) oauth2Complete(w http.ResponseWriter, r *http.Request) {
	authedUserID := r.Header.Get("Mattermost-User-ID")
	if authedUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}
	code := r.URL.Query().Get("code")
	if len(code) == 0 {
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}
	state := r.URL.Query().Get("state")

	storedState, appErr := o.api.KVGet(o.getStateKey(authedUserID))
	if appErr != nil {
		httputils.WriteUnauthorizedError(w, appErr)
		return
	}

	if string(storedState) != state {
		httputils.WriteUnauthorizedError(w, errors.New("state is different"))
		return
	}

	userID := strings.Split(state, "_")[1]
	if userID != authedUserID {
		httputils.WriteUnauthorizedError(w, errors.New("authed user is not the same as state user"))
		return
	}

	ctx := context.Background()
	tok, err := o.config.Exchange(ctx, code)
	if err != nil {
		httputils.WriteUnauthorizedError(w, err)
		return
	}

	rawToken, err := json.Marshal(tok)
	if err != nil {
		httputils.WriteUnauthorizedError(w, err)
		return
	}
	o.api.KVSet(o.getTokenKey(userID), rawToken)

	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
			<head>
				<script>
					window.close();
				</script>
			</head>
			<body>
				<p>%s</p>
			</body>
		</html>
		`, o.connectedString)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))

	if o.onConnect != nil {
		o.onConnect(userID, tok)
	}
}

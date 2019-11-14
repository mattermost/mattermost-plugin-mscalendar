// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-plugin-msoffice/server/store"
)

func (h *Handler) oauth2Complete(w http.ResponseWriter, r *http.Request) {
	authedUserID := r.Header.Get("Mattermost-User-ID")
	if authedUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	oconf := h.Remote.NewOAuth2Config(h.Config)

	code := r.URL.Query().Get("code")
	if len(code) == 0 {
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	err := h.OAuth2StateStore.VerifyOAuth2State(state)
	if err != nil {
		http.Error(w, "missing stored state: "+err.Error(), http.StatusBadRequest)
		return
	}

	mattermostUserID := strings.Split(state, "_")[1]
	if mattermostUserID != authedUserID {
		http.Error(w, "Not authorized, user ID mismatch.", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	tok, err := oconf.Exchange(ctx, code)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := h.Remote.NewClient(ctx, h.Config, tok, h.Logger)
	me, err := client.GetMe()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	u := &store.User{
		PluginVersion:    h.Config.PluginVersion,
		MattermostUserID: mattermostUserID,
		Remote:           me,
		OAuth2Token:      tok,
	}

	err = h.UserStore.StoreUser(u)
	if err != nil {
		http.Error(w, "Unable to connect: "+err.Error(), http.StatusInternalServerError)
		return
	}

	message := fmt.Sprintf("### Welcome to the Microsoft Office plugin!\n"+
		"Here is some info to prove we got you logged in\n"+
		"Name: %s \n", me.DisplayName)
	h.BotPoster.PostDirect(mattermostUserID, message, "custom_TODO")

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

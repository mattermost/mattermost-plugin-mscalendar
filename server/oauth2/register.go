// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package oauth2

import (
	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
)

// InitRouter initializes the router.
func RegisterHTTP(r *mux.Router) {
	oauth2Router := r.PathPrefix(config.PathOAuth2).Subrouter()
	oauth2Router.HandleFunc("/connect", oauth2Connect).Methods("GET")
	oauth2Router.HandleFunc(config.PathComplete, oauth2Complete).Methods("GET")
}

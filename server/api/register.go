// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
)

// InitRouter initializes the router.
func RegisterHTTP(r *mux.Router) {
	apiRouter := r.PathPrefix(config.PathAPI).Subrouter()
	apiRouter.HandleFunc("/authorized", apiGetAuthorized).Methods("GET")

	// TODO Refactor this to api/notification.go, remove consts
	notificationRouter := r.PathPrefix(config.PathNotification).Subrouter()
	notificationRouter.HandleFunc(config.PathEvent, notification).Methods("POST")

	postActionRouter := r.PathPrefix(config.PathPostAction).Subrouter()
	postActionRouter.HandleFunc(config.PathAccept, postActionAccept).Methods("POST")
	postActionRouter.HandleFunc(config.PathDecline, postActionDecline).Methods("POST")
	postActionRouter.HandleFunc(config.PathTentative, postActionTentative).Methods("POST")
	postActionRouter.HandleFunc(config.PathRespond, postActionRespond).Methods("POST")
}

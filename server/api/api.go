// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
)

type api struct {
	mscalendar.Env
	mscalendar.NotificationProcessor
}

// Init initializes the router.
func Init(h *httputils.Handler, env mscalendar.Env, notificationProcessor mscalendar.NotificationProcessor) {
	api := &api{
		Env:                   env,
		NotificationProcessor: notificationProcessor,
	}

	apiRouter := h.Router.PathPrefix(config.PathAPI).Subrouter()
	apiRouter.HandleFunc("/authorized", api.getAuthorized).Methods("GET")

	notificationRouter := h.Router.PathPrefix(config.PathNotification).Subrouter()
	notificationRouter.HandleFunc(config.PathEvent, api.notification).Methods("POST")

	postActionRouter := h.Router.PathPrefix(config.PathPostAction).Subrouter()
	postActionRouter.HandleFunc(config.PathAccept, api.postActionAccept).Methods("POST")
	postActionRouter.HandleFunc(config.PathDecline, api.postActionDecline).Methods("POST")
	postActionRouter.HandleFunc(config.PathTentative, api.postActionTentative).Methods("POST")
	postActionRouter.HandleFunc(config.PathRespond, api.postActionRespond).Methods("POST")
	postActionRouter.HandleFunc(config.PathConfirmStatusChange, api.postActionConfirmStatusChange).Methods("POST")

	dialogRouter := h.Router.PathPrefix(config.PathAutocomplete).Subrouter()
	dialogRouter.HandleFunc(config.PathUsers, api.autocompleteConnectedUsers)
	dialogRouter.HandleFunc(config.PathChannels, api.autocompleteChannels)

	apiRoutes := h.Router.PathPrefix(config.InternalAPIPath).Subrouter()
	eventsRouter := apiRoutes.PathPrefix(config.PathEvents).Subrouter()
	eventsRouter.HandleFunc(config.PathCreate, api.createEvent).Methods("POST")

	notificationRouter.HandleFunc("/{fname}", func(w http.ResponseWriter, r *http.Request) {
		if api.GoogleDomainVerifyKey == "" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Domain verify key is not set"))
			return
		}

		parts := strings.Split(r.URL.Path, "/")
		fname := parts[len(parts)-1]
		if fname != api.GoogleDomainVerifyKey {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Incorrect file name requested"))
			return
		}

		resp := "google-site-verification: " + api.GoogleDomainVerifyKey
		w.Write([]byte(resp))
	})
}

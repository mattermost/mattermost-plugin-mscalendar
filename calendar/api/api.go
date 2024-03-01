// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/engine"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/httputils"
)

type api struct {
	engine.Env
	engine.NotificationProcessor
}

// Init initializes the router.
func Init(h *httputils.Handler, env engine.Env, notificationProcessor engine.NotificationProcessor) {
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

	apiRoutes := h.Router.PathPrefix(config.InternalAPIPath).Subrouter()
	eventsRouter := apiRoutes.PathPrefix(config.PathEvents).Subrouter()
	eventsRouter.HandleFunc(config.PathCreate, api.createEvent).Methods("POST")
	apiRoutes.HandleFunc(config.PathConnectedUser, api.connectedUserHandler)

	// Returns provider information for the plugin to use
	apiRoutes.HandleFunc(config.PathProvider, func(w http.ResponseWriter, r *http.Request) {
		httputils.WriteJSONResponse(w, config.Provider, 200)
	})
}

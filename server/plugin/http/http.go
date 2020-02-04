// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
)

// Handler is an http.Handler for all plugin HTTP endpoints
type Handler struct {
	// Logger utils.Logger
	*config.Config
	*mux.Router
}

// InitRouter initializes the router.
func NewHandler(conf *config.Config) *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
		Config: conf,
	}

	apiRouter := h.Router.PathPrefix(config.PathAPI).Subrouter()
	apiRouter.HandleFunc("/authorized", h.apiGetAuthorized).Methods("GET")

	// TODO Refactor this to api/notification.go, remove consts
	notificationRouter := h.Router.PathPrefix(config.PathNotification).Subrouter()
	notificationRouter.HandleFunc(config.PathEvent, h.notification).Methods("POST")

	actionRouter := h.Router.PathPrefix(config.PathPostAction).Subrouter()
	actionRouter.HandleFunc(config.PathAccept, h.actionAccept).Methods("POST")
	actionRouter.HandleFunc(config.PathDecline, h.actionDecline).Methods("POST")
	actionRouter.HandleFunc(config.PathTentative, h.actionTentative).Methods("POST")
	actionRouter.HandleFunc(config.PathRespond, h.actionRespond).Methods("POST")

	oauth2Router := h.Router.PathPrefix(config.PathOAuth2).Subrouter()
	oauth2Router.HandleFunc(config.PathConnect, h.oauth2Connect).Methods("GET")
	oauth2Router.HandleFunc(config.PathComplete, h.oauth2Complete).Methods("GET")

	h.Router.Handle("{anything:.*}", http.NotFoundHandler())
	return h
}

func (h *Handler) jsonError(w http.ResponseWriter, statusCode int, summary string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	b, _ := json.Marshal(struct {
		Error   string `json:"error"`
		Summary string `json:"details"`
	}{
		Summary: summary,
		// Summary:   "An internal error has occurred. Check app server logs for details.",
		Error: err.Error(),
	})
	_, _ = w.Write(b)
}

func (h *Handler) internalServerError(w http.ResponseWriter, err error) {
	h.jsonError(w, http.StatusInternalServerError, "An internal error has occurred. Check app server logs for details.", err)
}

func (h *Handler) badRequest(w http.ResponseWriter, err error) {
	h.jsonError(w, http.StatusBadRequest, "Invalid request.", err)
}

func (h *Handler) notFound(w http.ResponseWriter, err error) {
	h.jsonError(w, http.StatusNotFound, "Not found.", err)
}

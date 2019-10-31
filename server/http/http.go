// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-server/model"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
)

// Handler is an http.Handler for interacting with workflows through a REST API
type Handler struct {
	Config *config.Config
	root   *mux.Router
}

// NewHandler constructs a new handler.
func NewProtoHandler() *Handler {
	h := &Handler{
		root: mux.NewRouter(),
	}

	api := h.root.PathPrefix("/api/v1").Subrouter()
	api.Use(authorizationRequired)
	api.HandleFunc("/authorized", h.getAuthorized).Methods("GET")

	user := h.root.PathPrefix("/user").Subrouter()
	user.Use(authorizationRequired)

	h.root.Handle("{anything:.*}", http.NotFoundHandler())
	return h
}

// CloneWithConfig creates a clone to use for handling a single request,
// with the current context
func (h *Handler) CloneWithConfig(conf *config.Config) *Handler {
	hh := *h
	hh.Config = conf
	return &hh
}

func (h *Handler) jsonError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	b, _ := json.Marshal(struct {
		Error   string `json:"error"`
		Details string `json:"details"`
	}{
		Error:   "An internal error has occurred. Check app server logs for details.",
		Details: err.Error(),
	})
	_, _ = w.Write(b)
}

func authorizationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID != "" {
			next.ServeHTTP(w, r)
			return
		}
		http.Error(w, "Not authorized", http.StatusUnauthorized)
	})
}

func (h *Handler) adminAuthorizationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		authorized, err := h.Config.IsAuthorizedAdmin(userID)
		if err != nil {
			h.Config.PAPI.LogError("Admin authorization failed", "error", err.Error())
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}
		if authorized {
			next.ServeHTTP(w, r)
			return
		}
		http.Error(w, "Not authorized", http.StatusUnauthorized)
	})
}

func (h *Handler) SendEphemeralPost(channelID, userID, message string) {
	ephemeralPost := &model.Post{
		ChannelId: channelID,
		UserId:    h.Config.BotUserId,
		Message:   message,
	}
	_ = h.Config.PAPI.SendEphemeralPost(userID, ephemeralPost)
}

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.root.ServeHTTP(w, r)
}

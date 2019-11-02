// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	"github.com/mattermost/mattermost-plugin-msoffice/server/user"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
)

// Handler is an http.Handler for interacting with workflows through a REST API
type Handler struct {
	Config            *config.Config
	UserStore         user.Store
	API               plugin.API
	BotPoster         utils.BotPoster
	IsAuthorizedAdmin func(userId string) (bool, error)
	root              *mux.Router
}

// InitRouter initializes the router.
func (h *Handler) InitRouter() {
	h.root = mux.NewRouter()
	api := h.root.PathPrefix("/api/v1").Subrouter()
	api.Use(authorizationRequired)
	api.HandleFunc("/authorized", h.apiGetAuthorized).Methods("GET")

	oauth2 := h.root.PathPrefix("/oauth2").Subrouter()
	oauth2.Use(authorizationRequired)
	oauth2.HandleFunc("/connect", h.oauth2Connect).Methods("GET")
	oauth2.HandleFunc("/complete", h.oauth2Complete).Methods("GET")

	h.root.Handle("{anything:.*}", http.NotFoundHandler())
	return
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
		authorized, err := h.IsAuthorizedAdmin(userID)
		if err != nil {
			h.API.LogError("Admin authorization failed", "error", err.Error())
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
	_ = h.API.SendEphemeralPost(userID, ephemeralPost)
}

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.root.ServeHTTP(w, r)
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/store"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/bot"
)

// Handler is an http.Handler for all plugin HTTP endpoints
type Handler struct {
	Config            *config.Config
	UserStore         store.UserStore
	OAuth2StateStore  store.OAuth2StateStore
	SubscriptionStore store.SubscriptionStore
	Logger            utils.Logger
	Poster            bot.Poster
	IsAuthorizedAdmin func(userId string) (bool, error)
	Remote            remote.Remote
	root              *mux.Router
}

// InitRouter initializes the router.
func (h *Handler) InitRouter() {
	h.root = mux.NewRouter()
	api := h.root.PathPrefix(config.APIPath).Subrouter()
	api.Use(h.authorizationRequired)
	api.HandleFunc("/authorized", h.apiGetAuthorized).Methods("GET")

	webhook := h.root.PathPrefix(config.WebhookPath).Subrouter()
	webhook.HandleFunc(config.WebhookEventPath, h.webhookEvent).Methods("POST")

	oauth2 := h.root.PathPrefix(config.OAuth2Path).Subrouter()
	oauth2.Use(h.authorizationRequired)
	oauth2.HandleFunc("/connect", h.oauth2Connect).Methods("GET")
	oauth2.HandleFunc(config.OAuth2CompletePath, h.oauth2Complete).Methods("GET")

	h.root.Handle("{anything:.*}", http.NotFoundHandler())
	return
}

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.root.ServeHTTP(w, r)
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

func (h *Handler) authorizationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID != "" {
			next.ServeHTTP(w, r)
			return
		}
		h.Logger.LogInfo("Not authorised", "userID", userID)
		http.Error(w, "Not authorized", http.StatusUnauthorized)
	})
}

func (h *Handler) adminAuthorizationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		authorized, err := h.IsAuthorizedAdmin(userID)
		if err != nil {
			h.Logger.LogError("Admin authorization failed", "error", err.Error())
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

package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
	"github.com/mattermost/mattermost-server/v6/model"
)

func (api *api) autocompleteUsers(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-Id")
	_, err := api.Store.LoadUser(mattermostUserID)
	if mattermostUserID == "" || errors.Is(err, store.ErrNotFound) {
		httputils.WriteUnauthorizedError(w, fmt.Errorf("unauthorized"))
		return
	}

	searchString := r.URL.Query().Get("search")
	results, err := api.Store.SearchInUserIndex(searchString, 10)
	if err != nil {
		utils.SlackAttachmentError(w, "unable to search in user index: "+err.Error())
		httputils.WriteInternalServerError(w, err)
		return
	}

	if err := httputils.WriteJSONResponse(w, results.ToDTO(), http.StatusOK); err != nil {
		api.Logger.With(bot.LogContext{"err": err.Error()}).Errorf("error sending response to user")
		httputils.WriteInternalServerError(w, err)
	}
}

type autocompleteChannel struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

func (api *api) autocompleteChannels(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-Id")
	_, err := api.Store.LoadUser(mattermostUserID)
	if mattermostUserID == "" || errors.Is(err, store.ErrNotFound) {
		httputils.WriteUnauthorizedError(w, fmt.Errorf("unauthorized"))
		return
	}

	searchString := r.URL.Query().Get("search")

	teams, err := api.PluginAPI.GetMattermostUserTeams(mattermostUserID)
	if err != nil {
		utils.SlackAttachmentError(w, "error getting user teams: "+err.Error())
		httputils.WriteInternalServerError(w, err)
		return
	}

	var channels []*model.Channel

	for _, team := range teams {
		teamChannels, err := api.PluginAPI.SearchLinkableChannelForUser(team.Id, mattermostUserID, searchString)
		if err != nil {
			utils.SlackAttachmentError(w, "error searching channels: "+err.Error())
			httputils.WriteInternalServerError(w, err)
			return
		}

		channels = append(channels, teamChannels...)
	}

	var chResponse []autocompleteChannel

	for _, ch := range channels {
		chResponse = append(chResponse, autocompleteChannel{ID: ch.Id, DisplayName: ch.DisplayName})
	}

	if err := httputils.WriteJSONResponse(w, chResponse, http.StatusOK); err != nil {
		api.Logger.With(bot.LogContext{"err": err.Error()}).Errorf("error sending response to user")
		httputils.WriteInternalServerError(w, err)
	}
}

package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
)

func (api *api) autocompleteUsers(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-Id")
	_, err := api.Store.LoadUser(mattermostUserID)
	if mattermostUserID == "" || errors.Is(err, store.ErrNotFound) {
		httputils.WriteUnauthorizedError(w, fmt.Errorf("unauthorized"))
		return
	}

	searchString := r.URL.Query().Get("search")
	result, err := api.Store.SearchInUserIndex(searchString, 10)
	if err != nil {
		utils.SlackAttachmentError(w, "unable to search in user index: "+err.Error())
		httputils.WriteInternalServerError(w, err)
		return
	}

	if err := httputils.WriteJSONResponse(w, result, http.StatusOK); err != nil {
		api.Logger.With(bot.LogContext{"err": err.Error()}).Errorf("error sending response to user")
		httputils.WriteInternalServerError(w, err)
	}
}

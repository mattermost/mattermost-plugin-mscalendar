package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/httputils"
)

func (api *api) connectedUserHandler(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-Id")
	if mattermostUserID == "" {
		httputils.WriteUnauthorizedError(w, fmt.Errorf("unauthorized"))
		return
	}

	_, errStore := api.Store.LoadUser(mattermostUserID)
	if errStore != nil && !errors.Is(errStore, store.ErrNotFound) {
		httputils.WriteInternalServerError(w, errStore)
		return
	}
	if errors.Is(errStore, store.ErrNotFound) {
		httputils.WriteUnauthorizedError(w, fmt.Errorf("unauthorized"))
		return
	}

	w.Write([]byte(`{"is_connected": true}`))
}

package settingspanel

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
	"github.com/mattermost/mattermost-server/v5/model"
)

type handler struct {
	panel Panel
}

func Init(h *httputils.Handler, panel Panel) {
	sh := &handler{
		panel: panel,
	}

	panelRouter := h.Router.PathPrefix("/").Subrouter()
	panelRouter.HandleFunc(panel.URL(), sh.handleAction).Methods("POST")
}

func (sh *handler) handleAction(w http.ResponseWriter, r *http.Request) {
	mattermostUserID := r.Header.Get("Mattermost-User-ID")
	if mattermostUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	id := ""
	value := ""

	for _, s := range sh.panel.GetSettingIDs() {
		id = s
		value = r.URL.Query().Get(s)
		if value != "" {
			break
		}
	}

	if value == "" {
		http.Error(w, "valid key not found", http.StatusBadRequest)
		return
	}

	sh.panel.Set(mattermostUserID, id, value)

	response := model.PostActionIntegrationResponse{}
	post, err := sh.panel.GetUpdatePost(mattermostUserID)
	if err == nil {
		response.Update = post
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response.ToJson())
}

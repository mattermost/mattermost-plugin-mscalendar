package settingspanel

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	ContextIDKey          = "setting_id"
	ContextButtonValueKey = "button_value"
	ContextOptionValueKey = "selected_option"
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

	request := model.PostActionIntegrationRequestFromJson(r.Body)
	if request == nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	id, ok := request.Context[ContextIDKey]
	if !ok {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	value, ok := request.Context[ContextButtonValueKey]
	if !ok {
		value, ok = request.Context[ContextOptionValueKey]
		if !ok {
			http.Error(w, "valid key not found", http.StatusBadRequest)
			return
		}
	}

	idString := id.(string)
	valueString := value.(string)
	sh.panel.Set(mattermostUserID, idString, valueString)

	response := model.PostActionIntegrationResponse{}
	post, err := sh.panel.GetUpdatePost(mattermostUserID)
	if err == nil {
		response.Update = post
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response.ToJson())
}

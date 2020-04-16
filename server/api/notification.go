// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/httputils"
)

func (api *api) notification(w http.ResponseWriter, req *http.Request) {
	err := api.NotificationProcessor.Enqueue(
		api.Env.Remote.HandleWebhook(w, req)...)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}
}

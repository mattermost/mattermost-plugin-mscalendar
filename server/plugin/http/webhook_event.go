// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package http

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-msoffice/server/api"
)

func (h *Handler) webhookEvent(w http.ResponseWriter, req *http.Request) {
	api := api.FromContext(req.Context())
	api.HandleEventNotification(w, req)
	return
}

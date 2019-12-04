// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package http

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-msoffice/server/api"
)

func (h *Handler) notification(w http.ResponseWriter, req *http.Request) {
	handler := api.NotificationHandlerFromContext(req.Context())
	handler.ServeHTTP(w, req)
	return
}

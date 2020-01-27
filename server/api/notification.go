// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
)

func notification(w http.ResponseWriter, req *http.Request) {
	handler := mscalendar.NotificationHandlerFromContext(req.Context())
	handler.ServeHTTP(w, req)
	return
}

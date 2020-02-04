// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"net/http"
)

func (api *api) notification(w http.ResponseWriter, req *http.Request) {
	api.NotificationProcessor.Enqueue(
		api.Env.Remote.HandleWebhook(w, req)...)
}

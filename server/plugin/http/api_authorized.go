// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package http

import "net/http"

func (h *Handler) apiGetAuthorized(w http.ResponseWriter, r *http.Request) {
	// if we've made it here, we're authorized.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"is_authorized": true}`))
}

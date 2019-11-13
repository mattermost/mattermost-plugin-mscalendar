// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"encoding/json"

	"github.com/mattermost/mattermost-plugin-msoffice/server/msgraph"
)

func (h *Handler) viewCalendar(parameters ...string) (string, error) {
	user, err := h.loadRemoteUser()
	if err != nil {
		return "", err
	}

	msgraphClient := msgraph.NewClient(h.Config, user.OAuth2Token)
	cals, err := msgraphClient.GetUserCalendar("")
	if err != nil {
		return "", err
	}

	bb, _ := json.MarshalIndent(cals, "", "  ")
	resp := "<><>" + string(bb)
	return resp, nil
}

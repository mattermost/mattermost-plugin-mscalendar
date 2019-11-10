// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"context"
	"encoding/json"
)

func (h *Handler) viewCalendar(parameters ...string) (string, error) {
	user, err := h.loadRemoteUser()
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	client := h.Remote.NewClient(ctx, h.Config, user.OAuth2Token)
	cals, err := client.GetUserCalendars("")
	if err != nil {
		return "", err
	}

	bb, _ := json.MarshalIndent(cals, "", "  ")
	resp := "<><>" + string(bb)
	return resp, nil
}

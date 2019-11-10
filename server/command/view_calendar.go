// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"context"
	"encoding/json"
	"time"
)

func (h *Handler) viewCalendar(parameters ...string) (string, error) {
	user, err := h.loadRemoteUser()
	if err != nil {
		return "", err
	}

	client := h.Remote.NewClient(context.Background(), h.Config, user.OAuth2Token, h.Logger)

	events, err := client.GetUserDefaultCalendarView(user.Remote.ID, time.Now(), time.Now().Add(14*24*time.Hour))
	if err != nil {
		return "", err
	}

	resp := ""
	for _, e := range events {
		bb, _ := json.MarshalIndent(e, "", "  ")
		resp += "  - ```\n" + string(bb) + "\n```\n"
	}

	return resp, nil
}

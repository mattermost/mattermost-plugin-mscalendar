// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"context"
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
)

func (h *Handler) viewCalendar(parameters ...string) (string, error) {
	user, err := h.UserStore.LoadUser(h.MattermostUserID)
	if err != nil {
		return "", err
	}

	client := h.Remote.NewClient(context.Background(), user.OAuth2Token)

	events, err := client.GetUserDefaultCalendarView(user.Remote.ID, time.Now(), time.Now().Add(14*24*time.Hour))
	if err != nil {
		return "", err
	}

	resp := ""
	for _, e := range events {
		resp += "  - " + e.ID + utils.JSONBlock(e)
	}

	return resp, nil
}

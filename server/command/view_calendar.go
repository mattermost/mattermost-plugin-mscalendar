// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"context"
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
)

func (h *Handler) viewCalendar(parameters ...string) (string, error) {
	user, err := h.UserStore.LoadUser(h.MattermostUserId)
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
		resp += "  - " + e.ID + "\n```\n" + utils.PrettyJSON(e) + "\n```\n"
	}

	return resp, nil
}

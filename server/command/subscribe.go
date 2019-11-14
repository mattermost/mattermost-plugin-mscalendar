// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"context"
	"fmt"

	"github.com/mattermost/mattermost-plugin-msoffice/server/store"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
)

func (h *Handler) subscribe(parameters ...string) (string, error) {
	user, err := h.UserStore.LoadUser(h.MattermostUserId)
	if err != nil {
		return "", err
	}

	switch {
	case len(parameters) == 0:
		client := h.Remote.NewClient(context.Background(), h.Config, user.OAuth2Token, h.Logger)
		sub, err := client.CreateUserEventSubscription(user.Remote.ID)
		if err != nil {
			return "", err
		}

		storedSub := &store.Subscription{
			Remote:              sub,
			MattermostCreatorID: h.MattermostUserId,
			PluginVersion:       h.Config.PluginVersion,
		}
		err = h.SubscriptionStore.StoreUserSubscription(user, storedSub)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Subscription %s created.", sub.ID), nil

	case len(parameters) == 1 && parameters[0] == "show":
		return fmt.Sprintf("Subscription: %s", utils.PrettyJSON(user.Settings.EventSubscriptionID)), nil

	case len(parameters) == 2 && parameters[0] == "delete":
		client := h.Remote.NewClient(context.Background(), h.Config, user.OAuth2Token, h.Logger)
		err := client.DeleteEventSubscription(parameters[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Subscription %s deleted.", parameters[1]), nil
	}

	return "bad syntax", nil
}

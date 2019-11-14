// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"context"
	"fmt"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	"github.com/mattermost/mattermost-plugin-msoffice/server/store"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
	"github.com/pkg/errors"
)

func (h *Handler) subscribe(parameters ...string) (string, error) {
	user, err := h.UserStore.LoadUser(h.MattermostUserID)
	if err != nil {
		return "", err
	}

	switch {
	case len(parameters) == 0:
		client := h.Remote.NewClient(context.Background(), user.OAuth2Token)
		notificationURL := h.Config.PluginURL + config.EventWebhookFullPath
		sub, err := client.CreateUserEventSubscription(user.Remote.ID, notificationURL)
		if err != nil {
			return "", err
		}

		storedSub := &store.Subscription{
			Remote:              sub,
			MattermostCreatorID: h.MattermostUserID,
			PluginVersion:       h.Config.PluginVersion,
		}
		err = h.SubscriptionStore.StoreUserSubscription(user, storedSub)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Subscription %s created.", sub.ID), nil

	case len(parameters) == 1 && parameters[0] == "show":
		sub, err := h.SubscriptionStore.LoadSubscription(user.Settings.EventSubscriptionID)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Subscription:\n%s", utils.PrettyJSON(sub)), nil

	case len(parameters) == 1 && parameters[0] == "renew":
		sub, err := h.SubscriptionStore.LoadSubscription(user.Settings.EventSubscriptionID)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Subscription:\n%s", utils.PrettyJSON(sub)), nil

	case len(parameters) >= 1 && parameters[0] == "delete":
		subscriptionID := ""
		if len(parameters) > 1 {
			subscriptionID = parameters[1]
		} else {
			subscriptionID = user.Settings.EventSubscriptionID
		}
		if subscriptionID == "" {
			return "", errors.New("no subscription specified")
		}
		client := h.Remote.NewClient(context.Background(), user.OAuth2Token)
		err := client.DeleteEventSubscription(subscriptionID)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Subscription %s deleted.", parameters[1]), nil
	}

	return "bad syntax", nil
}

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
		// sub, err := client.CreateEventSubscription(
		sub, err := client.CreateEventMessageSubscription(
			h.Config.PluginURL + config.EventNotificationFullPath)
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
		return fmt.Sprintf("Subscription:%s", utils.JSONBlock(sub)), nil

	case len(parameters) == 1 && parameters[0] == "renew":
		sub, err := h.SubscriptionStore.LoadSubscription(user.Settings.EventSubscriptionID)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Subscription:%s", utils.JSONBlock(sub)), nil

	case len(parameters) >= 1 && parameters[0] == "delete":
		subscriptionID := ""
		var updateUser *store.User
		if len(parameters) > 1 {
			subscriptionID = parameters[1]
		} else {
			subscriptionID = user.Settings.EventSubscriptionID
			updateUser = user
		}
		if subscriptionID == "" {
			return "", errors.New("subscription is not specified")
		}
		client := h.Remote.NewClient(context.Background(), user.OAuth2Token)
		err := client.DeleteSubscription(subscriptionID)
		if err != nil {
			return "", errors.WithMessagef(err, "failed to delete subscription %s", subscriptionID)
		}

		err = h.SubscriptionStore.DeleteUserSubscription(updateUser, subscriptionID)
		if err != nil {
			return "", errors.WithMessagef(err, "failed to delete subscription %s", subscriptionID)
		}

		return fmt.Sprintf("Subscription %s deleted.", subscriptionID), nil
	}
	return "bad syntax", nil
}

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/store"
	"github.com/pkg/errors"
)

func (api *api) CreateUserEventSubscription() (*store.Subscription, error) {
	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}
	// sub, err := client.CreateEventSubscription(
	sub, err := client.CreateEventMessageSubscription(
		api.Config.PluginURL + config.EventNotificationFullPath)
	if err != nil {
		return nil, err
	}

	storedSub := &store.Subscription{
		Remote:              sub,
		MattermostCreatorID: api.mattermostUserID,
		PluginVersion:       api.Config.PluginVersion,
	}
	err = api.SubscriptionStore.StoreUserSubscription(api.user, storedSub)
	if err != nil {
		return nil, err
	}

	return storedSub, nil
}

func (api *api) LoadUserEventSubscription() (*store.Subscription, error) {
	err := api.Filter(withUser)
	if err != nil {
		return nil, err
	}

	storedSub, err := api.SubscriptionStore.LoadSubscription(api.user.Settings.EventSubscriptionID)
	if err != nil {
		return nil, err
	}
	return storedSub, err
}

func (api *api) ListRemoteSubscriptions() ([]*remote.Subscription, error) {
	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}
	subs, err := client.ListSubscriptions()
	if err != nil {
		return nil, err
	}
	return subs, nil
}

func (api *api) RenewUserEventSubscription() (*store.Subscription, error) {
	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}

	subscriptionID := api.user.Settings.EventSubscriptionID
	expires, err := client.RenewSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}

	storedSub, err := api.SubscriptionStore.LoadSubscription(api.user.Settings.EventSubscriptionID)
	if err != nil {
		return nil, err
	}

	storedSub.Remote.ExpirationDateTime = expires.Format(time.RFC3339)

	err = api.SubscriptionStore.StoreUserSubscription(api.user, storedSub)
	if err != nil {
		return nil, err
	}
	return storedSub, err
}

func (api *api) DeleteUserEventSubscription() error {
	err := api.Filter(withUser)
	if err != nil {
		return err
	}
	subscriptionID := api.user.Settings.EventSubscriptionID

	err = api.SubscriptionStore.DeleteUserSubscription(api.user, subscriptionID)
	if err != nil {
		return errors.WithMessagef(err, "failed to delete subscription %s", subscriptionID)
	}

	err = api.DeleteOrphanedSubscription(subscriptionID)
	if err != nil {
		return err
	}
	return nil
}

func (api *api) DeleteOrphanedSubscription(subscriptionID string) error {
	client, err := api.MakeClient()
	if err != nil {
		return errors.WithMessagef(err, "failed to delete subscription %s", subscriptionID)
	}
	err = client.DeleteSubscription(subscriptionID)
	if err != nil {
		return errors.WithMessagef(err, "failed to delete subscription %s", subscriptionID)
	}
	return nil
}

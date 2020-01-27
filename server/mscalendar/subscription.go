// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
)

type Subscriptions interface {
	CreateUserEventSubscription() (*store.Subscription, error)
	RenewUserEventSubscription() (*store.Subscription, error)
	DeleteOrphanedSubscription(ID string) error
	DeleteUserEventSubscription() error
	ListRemoteSubscriptions() ([]*remote.Subscription, error)
	LoadUserEventSubscription() (*store.Subscription, error)
}

func (mscalendar *mscalendar) CreateUserEventSubscription() (*store.Subscription, error) {
	client, err := mscalendar.MakeClient()
	if err != nil {
		return nil, err
	}
	sub, err := client.CreateSubscription(
		mscalendar.Config.PluginURL + config.FullPathEventNotification)
	if err != nil {
		return nil, err
	}

	storedSub := &store.Subscription{
		Remote:              sub,
		MattermostCreatorID: mscalendar.mattermostUserID,
		PluginVersion:       mscalendar.Config.PluginVersion,
	}
	err = mscalendar.SubscriptionStore.StoreUserSubscription(mscalendar.user, storedSub)
	if err != nil {
		return nil, err
	}

	return storedSub, nil
}

func (mscalendar *mscalendar) LoadUserEventSubscription() (*store.Subscription, error) {
	err := mscalendar.Filter(withUser)
	if err != nil {
		return nil, err
	}

	storedSub, err := mscalendar.SubscriptionStore.LoadSubscription(mscalendar.user.Settings.EventSubscriptionID)
	if err != nil {
		return nil, err
	}
	return storedSub, err
}

func (mscalendar *mscalendar) ListRemoteSubscriptions() ([]*remote.Subscription, error) {
	client, err := mscalendar.MakeClient()
	if err != nil {
		return nil, err
	}
	subs, err := client.ListSubscriptions()
	if err != nil {
		return nil, err
	}
	return subs, nil
}

func (mscalendar *mscalendar) RenewUserEventSubscription() (*store.Subscription, error) {
	client, err := mscalendar.MakeClient()
	if err != nil {
		return nil, err
	}

	subscriptionID := mscalendar.user.Settings.EventSubscriptionID
	renewed, err := client.RenewSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}

	storedSub, err := mscalendar.SubscriptionStore.LoadSubscription(mscalendar.user.Settings.EventSubscriptionID)
	if err != nil {
		return nil, err
	}
	storedSub.Remote = renewed

	err = mscalendar.SubscriptionStore.StoreUserSubscription(mscalendar.user, storedSub)
	if err != nil {
		return nil, err
	}
	return storedSub, err
}

func (mscalendar *mscalendar) DeleteUserEventSubscription() error {
	err := mscalendar.Filter(withUser)
	if err != nil {
		return err
	}
	subscriptionID := mscalendar.user.Settings.EventSubscriptionID

	err = mscalendar.SubscriptionStore.DeleteUserSubscription(mscalendar.user, subscriptionID)
	if err != nil {
		return errors.WithMessagef(err, "failed to delete subscription %s", subscriptionID)
	}

	err = mscalendar.DeleteOrphanedSubscription(subscriptionID)
	if err != nil {
		return err
	}
	return nil
}

func (mscalendar *mscalendar) DeleteOrphanedSubscription(subscriptionID string) error {
	client, err := mscalendar.MakeClient()
	if err != nil {
		return err
	}
	err = client.DeleteSubscription(subscriptionID)
	if err != nil {
		return errors.WithMessagef(err, "failed to delete subscription %s", subscriptionID)
	}
	return nil
}

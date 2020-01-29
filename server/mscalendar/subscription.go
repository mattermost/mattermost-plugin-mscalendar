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
	CreateMyEventSubscription() (*store.Subscription, error)
	RenewMyEventSubscription() (*store.Subscription, error)
	DeleteOrphanedSubscription(ID string) error
	DeleteMyEventSubscription() error
	ListRemoteSubscriptions() ([]*remote.Subscription, error)
	LoadMyEventSubscription() (*store.Subscription, error)
}

func (mscalendar *mscalendar) CreateMyEventSubscription() (*store.Subscription, error) {
	err := mscalendar.Filter(
		withClient,
	)
	if err != nil {
		return nil, err
	}

	sub, err := mscalendar.client.CreateMySubscription(
		mscalendar.Config.PluginURL + config.FullPathEventNotification)
	if err != nil {
		return nil, err
	}

	storedSub := &store.Subscription{
		Remote:              sub,
		MattermostCreatorID: mscalendar.actingUser.MattermostUserID,
		PluginVersion:       mscalendar.Config.PluginVersion,
	}
	err = mscalendar.SubscriptionStore.StoreUserSubscription(mscalendar.actingUser.User, storedSub)
	if err != nil {
		return nil, err
	}

	return storedSub, nil
}

func (mscalendar *mscalendar) LoadMyEventSubscription() (*store.Subscription, error) {
	err := mscalendar.Filter(
		withActingUserExpanded,
	)
	if err != nil {
		return nil, err
	}

	storedSub, err := mscalendar.SubscriptionStore.LoadSubscription(mscalendar.actingUser.Settings.EventSubscriptionID)
	if err != nil {
		return nil, err
	}
	return storedSub, err
}

func (mscalendar *mscalendar) ListRemoteSubscriptions() ([]*remote.Subscription, error) {
	err := mscalendar.Filter(
		withClient,
	)
	if err != nil {
		return nil, err
	}

	subs, err := mscalendar.client.ListSubscriptions()
	if err != nil {
		return nil, err
	}
	return subs, nil
}

func (mscalendar *mscalendar) RenewMyEventSubscription() (*store.Subscription, error) {
	err := mscalendar.Filter(
		withClient,
	)
	if err != nil {
		return nil, err
	}

	subscriptionID := mscalendar.actingUser.Settings.EventSubscriptionID
	renewed, err := mscalendar.client.RenewSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}

	storedSub, err := mscalendar.SubscriptionStore.LoadSubscription(mscalendar.actingUser.Settings.EventSubscriptionID)
	if err != nil {
		return nil, err
	}
	storedSub.Remote = renewed

	err = mscalendar.SubscriptionStore.StoreUserSubscription(mscalendar.actingUser.User, storedSub)
	if err != nil {
		return nil, err
	}
	return storedSub, err
}

func (mscalendar *mscalendar) DeleteMyEventSubscription() error {
	err := mscalendar.Filter(
		withActingUserExpanded,
	)
	if err != nil {
		return err
	}

	subscriptionID := mscalendar.actingUser.Settings.EventSubscriptionID

	err = mscalendar.SubscriptionStore.DeleteUserSubscription(mscalendar.actingUser.User, subscriptionID)
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
	err := mscalendar.Filter(
		withClient,
	)
	if err != nil {
		return err
	}
	err = mscalendar.client.DeleteSubscription(subscriptionID)
	if err != nil {
		return errors.WithMessagef(err, "failed to delete subscription %s", subscriptionID)
	}
	return nil
}

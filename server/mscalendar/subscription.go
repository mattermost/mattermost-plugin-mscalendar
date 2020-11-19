// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"strings"

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

func (m *mscalendar) CreateMyEventSubscription() (*store.Subscription, error) {
	err := m.Filter(withClient)
	if err != nil {
		return nil, err
	}

	sub, err := m.client.CreateMySubscription(
		m.Config.PluginURL + config.FullPathEventNotification)
	if err != nil {
		return nil, err
	}

	storedSub := &store.Subscription{
		Remote:              sub,
		MattermostCreatorID: m.actingUser.MattermostUserID,
		PluginVersion:       m.Config.PluginVersion,
	}
	err = m.Store.StoreUserSubscription(m.actingUser.User, storedSub)
	if err != nil {
		return nil, err
	}

	return storedSub, nil
}

func (m *mscalendar) LoadMyEventSubscription() (*store.Subscription, error) {
	err := m.Filter(withActingUserExpanded)
	if err != nil {
		return nil, err
	}
	storedSub, err := m.Store.LoadSubscription(m.actingUser.Settings.EventSubscriptionID)
	if err != nil {
		return nil, err
	}
	return storedSub, err
}

func (m *mscalendar) ListRemoteSubscriptions() ([]*remote.Subscription, error) {
	err := m.Filter(withClient)
	if err != nil {
		return nil, err
	}
	subs, err := m.client.ListSubscriptions()
	if err != nil {
		return nil, err
	}
	return subs, nil
}

func (m *mscalendar) RenewMyEventSubscription() (*store.Subscription, error) {
	err := m.Filter(withClient)
	if err != nil {
		return nil, err
	}

	subscriptionID := m.actingUser.Settings.EventSubscriptionID
	if subscriptionID == "" {
		return nil, nil
	}
	renewed, err := m.client.RenewSubscription(subscriptionID)
	if err != nil {
		if strings.Contains(err.Error(), "The object was not found") {
			err = m.Store.DeleteUserSubscription(m.actingUser.User, subscriptionID)
			if err != nil {
				return nil, err
			}

			m.Logger.Infof("Subscription %s for Mattermost user %s has expired. Creating a new subscription now.", subscriptionID, m.actingUser.MattermostUserID)
			return m.CreateMyEventSubscription()
		}
		return nil, err
	}

	storedSub, err := m.Store.LoadSubscription(m.actingUser.Settings.EventSubscriptionID)
	if err != nil {
		return nil, err
	}
	storedSub.Remote = renewed

	err = m.Store.StoreUserSubscription(m.actingUser.User, storedSub)
	if err != nil {
		return nil, err
	}
	return storedSub, err
}

func (m *mscalendar) DeleteMyEventSubscription() error {
	err := m.Filter(withActingUserExpanded)
	if err != nil {
		return err
	}

	subscriptionID := m.actingUser.Settings.EventSubscriptionID

	err = m.Store.DeleteUserSubscription(m.actingUser.User, subscriptionID)
	if err != nil {
		return errors.WithMessagef(err, "failed to delete subscription %s", subscriptionID)
	}

	err = m.DeleteOrphanedSubscription(subscriptionID)
	if err != nil {
		return err
	}
	return nil
}

func (m *mscalendar) DeleteOrphanedSubscription(subscriptionID string) error {
	err := m.Filter(withClient)
	if err != nil {
		return err
	}
	err = m.client.DeleteSubscription(subscriptionID)
	if err != nil {
		return errors.WithMessagef(err, "failed to delete subscription %s", subscriptionID)
	}
	return nil
}

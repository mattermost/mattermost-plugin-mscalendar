// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"github.com/larkox/mattermost-plugin-utils/bot/logger"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/kvstore"
)

type SubscriptionStore interface {
	LoadSubscription(subscriptionID string) (*Subscription, error)
	StoreUserSubscription(user *User, subscription *Subscription) error
	DeleteUserSubscription(user *User, subscriptionID string) error
}

type Subscription struct {
	PluginVersion       string
	Remote              *remote.Subscription
	MattermostCreatorID string
}

func (s *pluginStore) LoadSubscription(subscriptionID string) (*Subscription, error) {
	sub := Subscription{}
	err := kvstore.LoadJSON(s.subscriptionKV, subscriptionID, &sub)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *pluginStore) StoreUserSubscription(user *User, subscription *Subscription) error {
	if user.Remote.ID != subscription.Remote.CreatorID {
		return errors.Errorf("User %q does not match the subscription creator %q",
			user.Remote.ID, subscription.Remote.CreatorID)
	}
	err := kvstore.StoreJSON(s.subscriptionKV, subscription.Remote.ID, subscription)
	if err != nil {
		return err
	}
	user.Settings.EventSubscriptionID = subscription.Remote.ID
	err = kvstore.StoreJSON(s.userKV, user.MattermostUserID, user)
	if err != nil {
		return err
	}

	s.Logger.With(logger.LogContext{
		"mattermostUserID": user.MattermostUserID,
		"remoteUserID":     subscription.Remote.CreatorID,
		"subscriptionID":   subscription.Remote.ID,
	}).Debugf("store: stored mattermost user subscription.")
	return nil
}

func (s *pluginStore) DeleteUserSubscription(user *User, subscriptionID string) error {
	err := s.subscriptionKV.Delete(subscriptionID)
	if err != nil {
		return err
	}
	mattermostUserID := ""
	if user != nil {
		user.Settings.EventSubscriptionID = ""
		err = s.StoreUser(user)
		if err != nil {
			return err
		}
		mattermostUserID = user.MattermostUserID
	}

	s.Logger.With(logger.LogContext{
		"mattermostUserID": mattermostUserID,
		"subscriptionID":   subscriptionID,
	}).Debugf("store: deleted mattermost user subscription.")
	return nil
}

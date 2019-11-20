// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"context"
	"net/http"
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/store"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/bot"
)

type OAuth2 interface {
	CompleteOAuth2(authedUserID, code, state string) error
	InitOAuth2(userID string) (url string, err error)
}

type Subscriptions interface {
	CreateUserEventSubscription() (*store.Subscription, error)
	RenewUserEventSubscription() (*store.Subscription, error)
	DeleteOrphanedSubscription(ID string) error
	DeleteUserEventSubscription() error
	LoadUserEventSubscription() (*store.Subscription, error)
	HandleEventNotification(w http.ResponseWriter, req *http.Request)
}

type Calendar interface {
	ViewCalendar(from, to time.Time) ([]*remote.Event, error)
}

type API interface {
	MakeClient() (remote.Client, error)

	OAuth2
	Subscriptions
	Calendar
}

// Dependencies contains all API dependencies
type Dependencies struct {
	UserStore         store.UserStore
	OAuth2StateStore  store.OAuth2StateStore
	SubscriptionStore store.SubscriptionStore
	Logger            utils.Logger
	Poster            bot.Poster
	Remote            remote.Remote
	IsAuthorizedAdmin func(userId string) (bool, error)
}

type api struct {
	*Dependencies
	*config.Config

	mattermostUserID string
	user             *store.User
}

func New(d Dependencies, c *config.Config, mattermostUserID string) API {
	return &api{
		Dependencies:     &d,
		Config:           c,
		mattermostUserID: mattermostUserID,
	}
}

type filterf func(*api) error

func (api *api) MakeClient() (remote.Client, error) {
	err := api.Filter(withUser)
	if err != nil {
		return nil, err
	}

	return api.Remote.NewClient(context.Background(), api.user.OAuth2Token), nil
}

func (api *api) Filter(filters ...filterf) error {
	for _, filter := range filters {
		err := filter(api)
		if err != nil {
			return err
		}
	}
	return nil
}

func withUser(api *api) error {
	if api.user != nil {
		return nil
	}

	user, err := api.UserStore.LoadUser(api.mattermostUserID)
	if err != nil {
		return err
	}

	api.user = user
	return nil
}

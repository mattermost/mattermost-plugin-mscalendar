// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"context"
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/store"
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
	ListRemoteSubscriptions() ([]*remote.Subscription, error)
	LoadUserEventSubscription() (*store.Subscription, error)
}

type Calendar interface {
	ViewCalendar(from, to time.Time) ([]*remote.Event, error)
	CreateEvent(event *remote.Event) (*remote.Event, error)
	CreateCalendar(calendarName string) (*remote.Calendar, error)
	DeleteCalendarByID(ID string) error
}

type Event interface {
	AcceptEvent(eventID string) error
	DeclineEvent(eventID string) error
	TentativelyAcceptEvent(eventID string) error
	RespondToEvent(eventID, response string) error
}

type Client interface {
	MakeClient() (remote.Client, error)
}

type API interface {
	Calendar
	Client
	Event
	OAuth2
	Subscriptions
}

// Dependencies contains all API dependencies
type Dependencies struct {
	UserStore         store.UserStore
	OAuth2StateStore  store.OAuth2StateStore
	SubscriptionStore store.SubscriptionStore
	EventStore        store.EventStore
	Logger            bot.Logger
	Poster            bot.Poster
	Remote            remote.Remote
	IsAuthorizedAdmin func(userId string) (bool, error)
}

type Config struct {
	*Dependencies
	*config.Config
}

type api struct {
	Config
	mattermostUserID string
	user             *store.User
}

func New(apiConfig Config, mattermostUserID string) API {
	return &api{
		Config:           apiConfig,
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

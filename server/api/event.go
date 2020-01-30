// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
)

func (api *api) CreateEvent(event *remote.Event, mattermostUserIDs []string) (*remote.Event, error) {
	// invite non-mapped Mattermost
	for id := range mattermostUserIDs {
		userID := mattermostUserIDs[id]
		_, err := api.UserStore.LoadUser(userID)
		if err != nil {
			if err.Error() == "not found" {
				err = api.Poster.DM(userID, "You have been invited to an MS office calendar event but have not linked your account.  Feel free to join us by connecting your www.office.com using `/msoffice connect`")
			}
		}
	}

	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}

	return client.CreateEvent(event)
}

func (api *api) AcceptEvent(eventID string) error {
	client, err := api.MakeClient()
	if err != nil {
		return err
	}

	return client.AcceptUserEvent(api.user.Remote.ID, eventID)
}

func (api *api) DeclineEvent(eventID string) error {
	client, err := api.MakeClient()
	if err != nil {
		return err
	}

	return client.DeclineUserEvent(api.user.Remote.ID, eventID)
}

func (api *api) TentativelyAcceptEvent(eventID string) error {
	client, err := api.MakeClient()
	if err != nil {
		return err
	}

	return client.TentativelyAcceptUserEvent(api.user.Remote.ID, eventID)
}

func (api *api) RespondToEvent(eventID, response string) error {
	if response == OptionNotResponded {
		return errors.New("Not responded is not a valid response")
	}

	client, err := api.MakeClient()
	if err != nil {
		return err
	}

	switch response {
	case OptionYes:
		return client.AcceptUserEvent(api.user.Remote.ID, eventID)
	case OptionNo:
		return client.DeclineUserEvent(api.user.Remote.ID, eventID)
	case OptionMaybe:
		return client.TentativelyAcceptUserEvent(api.user.Remote.ID, eventID)
	default:
		return errors.New(response + " is not a valid response")
	}
}

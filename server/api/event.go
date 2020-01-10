// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import "github.com/pkg/errors"

func (api *api) AcceptEvent(eventID string) error {
	client, err := api.NewClient()
	if err != nil {
		return err
	}

	return client.AcceptUserEvent(api.user.Remote.ID, eventID)
}

func (api *api) DeclineEvent(eventID string) error {
	client, err := api.NewClient()
	if err != nil {
		return err
	}

	return client.DeclineUserEvent(api.user.Remote.ID, eventID)
}

func (api *api) TentativelyAcceptEvent(eventID string) error {
	client, err := api.NewClient()
	if err != nil {
		return err
	}

	return client.TentativelyAcceptUserEvent(api.user.Remote.ID, eventID)
}

func (api *api) RespondToEvent(eventID, response string) error {
	if response == OptionNotResponded {
		return errors.New("Not responded is not a valid response")
	}

	client, err := api.NewClient()
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

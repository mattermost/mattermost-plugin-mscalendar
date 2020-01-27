// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import "github.com/pkg/errors"

type Event interface {
	AcceptEvent(eventID string) error
	DeclineEvent(eventID string) error
	TentativelyAcceptEvent(eventID string) error
	RespondToEvent(eventID, response string) error
}

func (mscalendar *mscalendar) AcceptEvent(eventID string) error {
	client, err := mscalendar.MakeClient()
	if err != nil {
		return err
	}

	return client.AcceptUserEvent(mscalendar.user.Remote.ID, eventID)
}

func (mscalendar *mscalendar) DeclineEvent(eventID string) error {
	client, err := mscalendar.MakeClient()
	if err != nil {
		return err
	}

	return client.DeclineUserEvent(mscalendar.user.Remote.ID, eventID)
}

func (mscalendar *mscalendar) TentativelyAcceptEvent(eventID string) error {
	client, err := mscalendar.MakeClient()
	if err != nil {
		return err
	}

	return client.TentativelyAcceptUserEvent(mscalendar.user.Remote.ID, eventID)
}

func (mscalendar *mscalendar) RespondToEvent(eventID, response string) error {
	if response == OptionNotResponded {
		return errors.New("Not responded is not a valid response")
	}

	client, err := mscalendar.MakeClient()
	if err != nil {
		return err
	}

	switch response {
	case OptionYes:
		return client.AcceptUserEvent(mscalendar.user.Remote.ID, eventID)
	case OptionNo:
		return client.DeclineUserEvent(mscalendar.user.Remote.ID, eventID)
	case OptionMaybe:
		return client.TentativelyAcceptUserEvent(mscalendar.user.Remote.ID, eventID)
	default:
		return errors.New(response + " is not a valid response")
	}
}

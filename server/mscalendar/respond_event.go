// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import "github.com/pkg/errors"

type EventResponder interface {
	AcceptEvent(user *User, eventID string) error
	DeclineEvent(user *User, eventID string) error
	TentativelyAcceptEvent(user *User, eventID string) error
	RespondToEvent(user *User, eventID, response string) error
}

func (mscalendar *mscalendar) AcceptEvent(user *User, eventID string) error {
	err := mscalendar.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return err
	}

	return mscalendar.client.AcceptEvent(user.Remote.ID, eventID)
}

func (mscalendar *mscalendar) DeclineEvent(user *User, eventID string) error {
	err := mscalendar.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return err
	}

	return mscalendar.client.DeclineEvent(user.Remote.ID, eventID)
}

func (mscalendar *mscalendar) TentativelyAcceptEvent(user *User, eventID string) error {
	err := mscalendar.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return err
	}

	return mscalendar.client.TentativelyAcceptEvent(user.Remote.ID, eventID)
}

func (mscalendar *mscalendar) RespondToEvent(user *User, eventID, response string) error {
	if response == OptionNotResponded {
		return errors.New("Not responded is not a valid response")
	}

	err := mscalendar.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return err
	}

	switch response {
	case OptionYes:
		return mscalendar.client.AcceptEvent(user.Remote.ID, eventID)
	case OptionNo:
		return mscalendar.client.DeclineEvent(user.Remote.ID, eventID)
	case OptionMaybe:
		return mscalendar.client.TentativelyAcceptEvent(user.Remote.ID, eventID)
	default:
		return errors.New(response + " is not a valid response")
	}
}

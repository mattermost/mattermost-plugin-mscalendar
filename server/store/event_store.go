// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"encoding/json"
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/kvstore"
)

// If event has an end date/time, its record will be set to expire ttlAfterEventEnd
// after its end. Events that have no end-date are created for defaultEventsTTL.
// Expirations are updated when events themselves are updated.
const ttlAfterEventEnd = 30 * 24 * time.Hour // 30 days
const defaultEventTTL = 30 * 24 * time.Hour  // 30 days

type Event struct {
	PluginVersion string
	Remote        *remote.Event
}

type EventStore interface {
	LoadUserEvent(userID, eventID string) (*Event, error)
	StoreUserEvent(userID string, event *Event) error
	DeleteUserEvent(userID, eventID string) error
}

func eventKey(userID, eventID string) string { return userID + "_" + eventID }

func (s *pluginStore) LoadUserEvent(userID, eventID string) (*Event, error) {
	event := Event{}
	err := kvstore.LoadJSON(s.eventKV, eventKey(userID, eventID), &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *pluginStore) StoreUserEvent(userID string, event *Event) error {
	now := time.Now()
	end := now.Add(defaultEventTTL)
	if event.Remote.End != nil {
		end = event.Remote.End.Time().Add(ttlAfterEventEnd)
		if end.Before(now) {
			// no point storing expired keys
			return nil
		}
	}

	ttl := int64(end.Sub(now).Seconds())
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	err = s.eventKV.StoreTTL(eventKey(userID, event.Remote.ID), data, ttl)
	if err != nil {
		return err
	}

	s.Logger.With(bot.LogContext{
		"UserID":  userID,
		"eventID": event.Remote.ID,
		"expires": end.String(),
	}).Debugf("store: stored user event.")

	return nil
}

func (s *pluginStore) DeleteUserEvent(userID, eventID string) error {
	err := s.eventKV.Delete(eventKey(userID, eventID))
	if err != nil {
		return err
	}

	s.Logger.With(bot.LogContext{
		"UserID":  userID,
		"eventID": eventID,
	}).Debugf("store: deleted event.")

	return nil
}

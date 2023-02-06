// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"encoding/json"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/kvstore"
)

// If event has an end date/time, its record will be set to expire ttlAfterEventEnd
// after its end. Events that have no end-date are created for defaultEventsTTL.
// Expirations are updated when events themselves are updated.
const ttlAfterEventEnd = 30 * 24 * time.Hour // 30 days
const defaultEventTTL = 30 * 24 * time.Hour  // 30 days

type Event struct {
	Remote        *remote.Event
	PluginVersion string
}

type EventStore interface {
	LoadUserEvent(mattermostUserID, eventID string) (*Event, error)
	StoreUserEvent(mattermostUserID string, event *Event) error
	DeleteUserEvent(mattermostUserID, eventID string) error
}

func eventKey(mattermostUserID, eventID string) string { return mattermostUserID + "_" + eventID }

func (s *pluginStore) LoadUserEvent(mattermostUserID, eventID string) (*Event, error) {
	event := Event{}
	err := kvstore.LoadJSON(s.eventKV, eventKey(mattermostUserID, eventID), &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *pluginStore) StoreUserEvent(mattermostUserID string, event *Event) error {
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
	err = s.eventKV.StoreTTL(eventKey(mattermostUserID, event.Remote.ICalUID), data, ttl)
	if err != nil {
		return err
	}

	s.Logger.With(bot.LogContext{
		"mattermostUserID": mattermostUserID,
		"eventID":          event.Remote.ID,
		"expires":          end.String(),
	}).Debugf("store: stored user event.")

	return nil
}

func (s *pluginStore) DeleteUserEvent(mattermostUserID, eventID string) error {
	err := s.eventKV.Delete(eventKey(mattermostUserID, eventID))
	if err != nil {
		return err
	}

	s.Logger.With(bot.LogContext{
		"mattermostUserID": mattermostUserID,
		"eventID":          eventID,
	}).Debugf("store: deleted event.")

	return nil
}

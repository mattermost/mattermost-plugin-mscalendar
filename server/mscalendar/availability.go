// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/views"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

const (
	availabilityTimeWindowSize      = 10 * time.Minute
	StatusSyncJobInterval           = 5 * time.Minute
	upcomingEventNotificationTime   = 10 * time.Minute
	upcomingEventNotificationWindow = (StatusSyncJobInterval * 9) / 10 //90% of the interval
)

type Availability interface {
	GetAvailabilities(users []*store.User) ([]*remote.ViewCalendarResponse, error)
	SyncStatus(mattermostUserID string) (string, error)
	SyncStatusAll() (string, error)
}

func (m *mscalendar) SyncStatus(mattermostUserID string) (string, error) {
	user, err := m.Store.LoadUserFromIndex(mattermostUserID)
	if err != nil {
		return "", err
	}

	return m.syncStatusUsers(store.UserIndex{user})
}

func (m *mscalendar) SyncStatusAll() (string, error) {
	isAdmin, err := m.IsAuthorizedAdmin(m.actingUser.MattermostUserID)
	if err != nil {
		return "", err
	}
	if !isAdmin {
		return "", errors.Errorf("Non-admin user attempting SyncStatusAll %s", m.actingUser.MattermostUserID)
	}
	err = m.Filter(withSuperuserClient)
	if err != nil {
		return "", err
	}

	userIndex, err := m.Store.LoadUserIndex()
	if err != nil {
		if err.Error() == "not found" {
			return "No users found in user index", nil
		}
		return "", err
	}

	return m.syncStatusUsers(userIndex)
}

func (m *mscalendar) syncStatusUsers(userIndex store.UserIndex) (string, error) {
	if len(userIndex) == 0 {
		return "No connected users found", nil
	}

	users := []*store.User{}
	for _, u := range userIndex {
		// TODO fetch users from kvstore in batches, and process in batches instead of all at once
		user, err := m.Store.LoadUser(u.MattermostUserID)
		if err != nil {
			return "", err
		}
		if user.Settings.UpdateStatus {
			users = append(users, user)
		}
	}
	if len(users) == 0 {
		return "No users want to have their status updated", nil
	}

	schedules, err := m.GetAvailabilities(users)
	if err != nil {
		return "", err
	}
	if len(schedules) == 0 {
		return "No schedule info found", nil
	}

	return m.setUserStatuses(users, schedules)
}

func (m *mscalendar) setUserStatuses(users []*store.User, schedules []*remote.ViewCalendarResponse) (string, error) {
	mattermostUserIDs := []string{}
	for _, u := range users {
		mattermostUserIDs = append(mattermostUserIDs, u.MattermostUserID)
	}

	statuses, appErr := m.PluginAPI.GetMattermostUserStatusesByIds(mattermostUserIDs)
	if appErr != nil {
		return "", appErr
	}
	statusMap := map[string]string{}
	for _, s := range statuses {
		statusMap[s.UserId] = s.Status
	}

	usersByRemoteID := map[string]*store.User{}
	for _, u := range users {
		usersByRemoteID[u.Remote.ID] = u
	}

	var res string
	for _, s := range schedules {
		user := usersByRemoteID[s.RemoteUserID]
		if s.Error != nil {
			m.Logger.Errorf("Error getting availability for %s: %s", user.Remote.Mail, s.Error.Message)
			continue
		}

		mattermostUserID := usersByRemoteID[s.RemoteUserID].MattermostUserID
		m.notifyUpcomingEvent(mattermostUserID, s.Events)
		status, ok := statusMap[mattermostUserID]
		if !ok {
			continue
		}

		var err error
		res, err = m.setStatusFromAvailability(user, status, s)
		if err != nil {
			m.Logger.Errorf("Error setting user %s status. %s", user.Remote.Mail, err.Error())
		}
	}
	if res != "" {
		return res, nil
	}

	return utils.JSONBlock(schedules), nil
}

func (m *mscalendar) setStatusFromAvailability(user *store.User, currentStatus string, res *remote.ViewCalendarResponse) (string, error) {
	events := filterBusyEvents(res.Events)

	if len(user.ActiveEvents) == 0 && len(events) == 0 {
		return "No events in local or remote. No status change.", nil
	}

	var message string
	if len(user.ActiveEvents) > 0 && len(events) == 0 {
		if currentStatus == model.STATUS_DND {
			err := m.setStatusOrAskUser(user, model.STATUS_ONLINE, events)
			if err != nil {
				return "", err
			}
			message = "User is no longer busy in calendar. Set status to online."
		} else {
			message = "User is no longer busy in calendar, but is not set to DND. No status change."
		}
		err := m.Store.StoreUserActiveEvents(user.MattermostUserID, []string{})
		if err != nil {
			return "", err
		}
		return message, nil
	}

	remoteHashes := []string{}
	for _, e := range events {
		h := fmt.Sprintf("%s %s", e.ID, e.Start.Time().UTC().Format(time.RFC3339))
		remoteHashes = append(remoteHashes, h)
	}

	if len(user.ActiveEvents) == 0 {
		if currentStatus == model.STATUS_DND {
			return "User was already marked as busy. No status change.", nil
		}
		err := m.setStatusOrAskUser(user, model.STATUS_DND, events)
		if err != nil {
			return "", err
		}
		err = m.Store.StoreUserActiveEvents(user.MattermostUserID, remoteHashes)
		if err != nil {
			return "", err
		}
		return "User was free, but is now busy. Set status to DND.", nil
	}

	newEventExists := false
	for _, r := range remoteHashes {
		found := false
		for _, loc := range user.ActiveEvents {
			if loc == r {
				found = true
				break
			}
		}
		if !found {
			newEventExists = true
			break
		}
	}

	if !newEventExists {
		return fmt.Sprintf("No change in active events. Total number of events: %d", len(events)), nil
	}

	if currentStatus != model.STATUS_DND {
		err := m.setStatusOrAskUser(user, model.STATUS_DND, events)
		if err != nil {
			return "", err
		}
		message = "User was free, but is now busy. Set status to DND."
	} else {
		message = "User is already busy. No status change."
	}
	err := m.Store.StoreUserActiveEvents(user.MattermostUserID, remoteHashes)
	if err != nil {
		return "", err
	}
	return message, nil
}

func (m *mscalendar) setStatusOrAskUser(user *store.User, status string, events []*remote.Event) error {
	if user.Settings.GetConfirmation {
		url := fmt.Sprintf("%s%s%s", m.Config.PluginURLPath, config.PathPostAction, config.PathConfirmStatusChange)
		_, err := m.Poster.DMWithAttachments(user.MattermostUserID, views.RenderStatusChangeNotificationView(events, status, url))
		return err
	}

	_, appErr := m.PluginAPI.UpdateMattermostUserStatus(user.MattermostUserID, status)
	if appErr != nil {
		return appErr
	}
	return nil
}

func (m *mscalendar) GetAvailabilities(users []*store.User) ([]*remote.ViewCalendarResponse, error) {
	err := m.Filter(withClient)
	if err != nil {
		return nil, err
	}

	start := time.Now().UTC()
	end := time.Now().UTC().Add(availabilityTimeWindowSize)

	params := []*remote.ViewCalendarParams{}
	for _, u := range users {
		params = append(params, &remote.ViewCalendarParams{
			RemoteUserID: u.Remote.ID,
			StartTime:    start,
			EndTime:      end,
		})
	}

	return m.client.DoBatchViewCalendarRequests(params)
}

func (m *mscalendar) notifyUpcomingEvent(mattermostUserID string, events []*remote.Event) {
	var timezone string
	for _, event := range events {
		upcomingTime := time.Now().Add(upcomingEventNotificationTime)
		start := event.Start.Time()
		diff := start.Sub(upcomingTime)

		if (diff < upcomingEventNotificationWindow) && (diff > -upcomingEventNotificationWindow) {
			var err error
			if timezone == "" {
				timezone, err = m.GetTimezoneByID(mattermostUserID)
				if err != nil {
					m.Logger.Errorf("notifyUpcomingEvent error getting timezone, err=%s", err.Error())
					return
				}
			}

			message, err := views.RenderScheduleItem(event, timezone)
			if err != nil {
				m.Logger.Errorf("notifyUpcomingEvent error rendering schedule item, err=", err.Error())
				continue
			}
			err = m.Poster.DM(mattermostUserID, message)
			if err != nil {
				m.Logger.Errorf("notifyUpcomingEvent error creating DM, err=", err.Error())
				continue
			}
		}
	}
}

func filterBusyEvents(events []*remote.Event) []*remote.Event {
	result := []*remote.Event{}
	for _, e := range events {
		if e.ShowAs == "busy" {
			result = append(result, e)
		}
	}
	return result
}

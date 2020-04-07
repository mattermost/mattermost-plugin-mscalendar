// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/views"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
)

const (
	availabilityTimeWindowSize      = 15
	StatusSyncJobInterval           = 5 * time.Minute
	upcomingEventNotificationTime   = 10 * time.Minute
	upcomingEventNotificationWindow = (StatusSyncJobInterval * 9) / 10 //90% of the interval
)

type Availability interface {
	GetAvailabilities(remoteUserID string, scheduleIDs []string) ([]*remote.ScheduleInformation, error)
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
	userIndex, err := m.Store.LoadUserIndex()
	if err != nil {
		if err.Error() == "not found" {
			return "No users found in user index", nil
		}
		return "", err
	}

	return m.syncStatusUsers(userIndex)
}

func (m *mscalendar) syncStatusUsers(users store.UserIndex) (string, error) {
	err := m.Filter(
		withClient,
		withUserExpanded(m.actingUser),
	)
	if err != nil {
		return "", err
	}

	if len(users) == 0 {
		return "No connected users found", nil
	}

	scheduleIDs := []string{}
	for _, u := range users {
		scheduleIDs = append(scheduleIDs, u.Email)
	}

	schedules, err := m.GetAvailabilities(m.actingUser.Remote.ID, scheduleIDs)
	if err != nil {
		return "", err
	}
	if len(schedules) == 0 {
		return "No schedule info found", nil
	}

	return m.setUserStatuses(users, schedules)
}

func (m *mscalendar) setUserStatuses(users store.UserIndex, schedules []*remote.ScheduleInformation) (string, error) {
	mattermostUserIDs := users.GetMattermostUserIDs()
	statuses, appErr := m.PluginAPI.GetMattermostUserStatusesByIds(mattermostUserIDs)
	if appErr != nil {
		return "", appErr
	}
	statusMap := map[string]string{}
	for _, s := range statuses {
		statusMap[s.UserId] = s.Status
	}

	usersByEmail := users.ByEmail()
	var res string
	for _, s := range schedules {
		if s.Error != nil {
			m.Logger.Errorf("Error getting availability for %s: %s", s.ScheduleID, s.Error.ResponseCode)
			continue
		}

		mattermostUserID := usersByEmail[s.ScheduleID].MattermostUserID

		m.notifyUpcomingEvent(mattermostUserID, s.ScheduleItems)
		status, ok := statusMap[mattermostUserID]
		if !ok {
			continue
		}

		res = m.setStatusFromAvailability(mattermostUserID, status, s.AvailabilityView)
	}
	if res != "" {
		return res, nil
	}

	return utils.JSONBlock(schedules), nil
}

func (m *mscalendar) notifyUpcomingEvent(mattermostUserID string, items []remote.ScheduleItem) {
	var timezone string
	for _, scheduleItem := range items {
		upcomingTime := time.Now().Add(upcomingEventNotificationTime)
		start := scheduleItem.Start.Time()
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

			message, err := views.RenderScheduleItem(scheduleItem, timezone)
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

func (m *mscalendar) GetAvailabilities(remoteUserID string, scheduleIDs []string) ([]*remote.ScheduleInformation, error) {
	err := m.Filter(withClient)
	if err != nil {
		return nil, err
	}

	start := remote.NewDateTime(time.Now().UTC(), "UTC")
	end := remote.NewDateTime(time.Now().UTC().Add(availabilityTimeWindowSize*time.Minute), "UTC")

	return m.client.GetSchedule(remoteUserID, scheduleIDs, start, end, availabilityTimeWindowSize)
}

func (m *mscalendar) setStatusFromAvailability(mattermostUserID, currentStatus string, av remote.AvailabilityView) string {
	currentAvailability := av[0]

	switch currentAvailability {
	case remote.AvailabilityViewFree:
		if currentStatus == "dnd" {
			m.PluginAPI.UpdateMattermostUserStatus(mattermostUserID, "online")
			return fmt.Sprintf("User is free. Setting user from %s to online.", currentStatus)
		} else {
			return fmt.Sprintf("User is free, and is already set to %s.", currentStatus)
		}
	case remote.AvailabilityViewTentative, remote.AvailabilityViewBusy:
		if currentStatus != "dnd" {
			m.PluginAPI.UpdateMattermostUserStatus(mattermostUserID, "dnd")
			return fmt.Sprintf("User is busy. Setting user from %s to dnd.", currentStatus)
		} else {
			return fmt.Sprintf("User is busy, and is already set to %s.", currentStatus)
		}
	case remote.AvailabilityViewOutOfOffice:
		if currentStatus != "offline" {
			m.PluginAPI.UpdateMattermostUserStatus(mattermostUserID, "offline")
			return fmt.Sprintf("User is out of office. Setting user from %s to offline", currentStatus)
		} else {
			return fmt.Sprintf("User is out of office, and is already set to %s.", currentStatus)
		}
	case remote.AvailabilityViewWorkingElsewhere:
		return fmt.Sprintf("User is working elsewhere. Pending implementation.")
	}

	return fmt.Sprintf("Availability view doesn't match %d", currentAvailability)
}

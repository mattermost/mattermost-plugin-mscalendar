// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"fmt"
	"sync"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/views"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
	"github.com/pkg/errors"
)

const (
	availabilityTimeWindowSize      = 10 // minutes
	StatusSyncJobInterval           = 5 * time.Minute
	upcomingEventNotificationTime   = 10 * time.Minute
	upcomingEventNotificationWindow = (StatusSyncJobInterval * 9) / 10 //90% of the interval
)

type Availability interface {
	GetAvailabilities(users []*store.User) ([]*remote.ScheduleInformation, error)
	SyncStatus(mattermostUserID string) (string, error)
	SyncStatusAll() (string, error)
}

func (m *mscalendar) SyncStatus(mattermostUserID string) (string, error) {
	user, err := m.Store.LoadUser(mattermostUserID)
	if err != nil {
		return "", err
	}
	if !user.Settings.UpdateStatus {
		return fmt.Sprintf("Your settings are set up to not update your status. You can change your settings using `/%s settings`", config.CommandTrigger), nil
	}

	return m.syncStatusUsers([]*store.User{user})
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

	statusSyncUsers := []*store.User{}
	reminderUsers := []*store.User{}
	for _, u := range userIndex {
		user, err := m.Store.LoadUser(u.MattermostUserID)
		if err != nil {
			return "", err
		}
		if user.Settings.UpdateStatus {
			statusSyncUsers = append(statusSyncUsers, user)
		}
		if user.Settings.GetReminders {
			reminderUsers = append(reminderUsers, user)
		}
	}

	w := sync.WaitGroup{}
	m.Logger.Debugf("%d user(s) want their status updated", len(statusSyncUsers))
	if len(statusSyncUsers) > 0 {
		w.Add(1)
		go func() {
			res, err := m.syncStatusUsers(statusSyncUsers)
			if err != nil {
				m.Logger.Errorf("Error syncing user statuses. Error: %v", err)
			}
			if res != "" {
				m.Logger.Debugf(res)
			}
			m.Logger.Debugf("Status sync finished")
			w.Done()
		}()
	}

	m.Logger.Debugf("%d user(s) want a reminder for upcoming events", len(reminderUsers))
	if len(reminderUsers) > 0 {
		w.Add(1)
		go func() {
			err := m.deliverReminders(reminderUsers)
			if err != nil {
				m.Logger.Errorf("Error delivering reminders. Error: %v", err)
			}
			m.Logger.Debugf("Reminder deliveries finished")
			w.Done()
		}()
	}

	w.Wait()
	return "", nil
}

func (m *mscalendar) deliverReminders(users []*store.User) error {
	err := m.Filter(withClient)
	if err != nil {
		return err
	}

	start := time.Now().UTC()
	end := time.Now().UTC().Add(availabilityTimeWindowSize * time.Minute)

	usersByRemoteID := map[string]*store.User{}
	params := []*remote.ViewCalendarParams{}
	for _, u := range users {
		params = append(params, &remote.ViewCalendarParams{
			RemoteUserID: u.Remote.ID,
			StartTime:    start,
			EndTime:      end,
		})
		usersByRemoteID[u.Remote.ID] = u
	}

	responses, err := m.client.DoBatchViewCalendarRequests(params)
	if err != nil {
		return err
	}
	for _, res := range responses {
		u := usersByRemoteID[res.RemoteUserID]
		m.notifyUpcomingEvent(u.MattermostUserID, res.Events)
	}

	return nil
}

func (m *mscalendar) syncStatusUsers(users []*store.User) (string, error) {
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

func (m *mscalendar) setUserStatuses(users []*store.User, schedules []*remote.ScheduleInformation) (string, error) {
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

	usersByEmail := map[string]*store.User{}
	for _, u := range users {
		usersByEmail[u.Remote.Mail] = u
	}

	var res string
	for _, s := range schedules {
		user := usersByEmail[s.ScheduleID]
		if s.Error != nil {
			m.Logger.Errorf("Error getting availability for %s: %s", user.Remote.Mail, s.Error.Message)
			continue
		}

		mattermostUserID := usersByEmail[s.ScheduleID].MattermostUserID
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

func (m *mscalendar) setStatusFromAvailability(user *store.User, currentStatus string, sched *remote.ScheduleInformation) (string, error) {
	if len(sched.AvailabilityView) == 0 {
		return "No availabilities to process", nil
	}
	currentAvailability := sched.AvailabilityView[0]

	switch currentAvailability {
	case remote.AvailabilityViewFree:
		if currentStatus == "dnd" {
			m.setStatusOrAskUser(user, "online", sched)
			return fmt.Sprintf("User is free. Setting user from %s to online.", currentStatus), nil
		} else {
			return fmt.Sprintf("User is free, and is already set to %s.", currentStatus), nil
		}
	case remote.AvailabilityViewBusy:
		if currentStatus != "dnd" {
			m.setStatusOrAskUser(user, "dnd", sched)
			return fmt.Sprintf("User is busy. Setting user from %s to dnd.", currentStatus), nil
		} else {
			return fmt.Sprintf("User is busy, and is already set to %s.", currentStatus), nil
		}
	case remote.AvailabilityViewOutOfOffice:
		if currentStatus != "offline" {
			m.setStatusOrAskUser(user, "offline", sched)
			return fmt.Sprintf("User is out of office. Setting user from %s to offline", currentStatus), nil
		} else {
			return fmt.Sprintf("User is out of office, and is already set to %s.", currentStatus), nil
		}
	case remote.AvailabilityViewWorkingElsewhere:
		return fmt.Sprintf("User is working elsewhere. Pending implementation."), nil
	case remote.AvailabilityViewTentative:
		return fmt.Sprintf("User's availability is tentative. Don't change status."), nil
	}

	return fmt.Sprintf("Availability view doesn't match %d", currentAvailability), nil
}

func (m *mscalendar) setStatusOrAskUser(user *store.User, status string, sched *remote.ScheduleInformation) error {
	if user.Settings.GetConfirmation {
		url := fmt.Sprintf("%s%s%s", m.Config.PluginURLPath, config.PathPostAction, config.PathConfirmStatusChange)
		_, err := m.Poster.DMWithAttachments(user.MattermostUserID, views.RenderStatusChangeNotificationView(sched, status, url))
		return err
	}

	_, appErr := m.PluginAPI.UpdateMattermostUserStatus(user.MattermostUserID, status)
	if appErr != nil {
		return appErr
	}
	return nil
}

func (m *mscalendar) GetAvailabilities(users []*store.User) ([]*remote.ScheduleInformation, error) {
	err := m.Filter(withClient)
	if err != nil {
		return nil, err
	}

	params := []*remote.ScheduleUserInfo{}
	for _, u := range users {
		params = append(params, &remote.ScheduleUserInfo{
			RemoteUserID: u.Remote.ID,
			Mail:         u.Remote.Mail,
		})
	}

	start := remote.NewDateTime(timeNowFunc().UTC(), "UTC")
	end := remote.NewDateTime(timeNowFunc().UTC().Add(availabilityTimeWindowSize*time.Minute), "UTC")

	return m.client.GetSchedule(params, start, end, availabilityTimeWindowSize)
}

func (m *mscalendar) notifyUpcomingEvent(mattermostUserID string, events []*remote.Event) {
	var timezone string
	for _, event := range events {
		if event.IsCancelled {
			continue
		}
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

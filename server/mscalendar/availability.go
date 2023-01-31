// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/views"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
)

const (
	calendarViewTimeWindowSize      = 10 * time.Minute
	StatusSyncJobInterval           = 5 * time.Minute
	upcomingEventNotificationTime   = 10 * time.Minute
	upcomingEventNotificationWindow = (StatusSyncJobInterval * 11) / 10 // 110% of the interval
	logTruncateMsg                  = "We've truncated the logs due to too many messages"
	logTruncateLimit                = 5
)

type Availability interface {
	GetCalendarViews(users []*store.User) ([]*remote.ViewCalendarResponse, error)
	Sync(mattermostUserID string) (string, error)
	SyncAll() (string, error)
}

func (m *mscalendar) Sync(mattermostUserID string) (string, error) {
	user, err := m.Store.LoadUserFromIndex(mattermostUserID)
	if err != nil {
		return "", err
	}

	return m.syncUsers(store.UserIndex{user})
}

func (m *mscalendar) SyncAll() (string, error) {
	err := m.Filter(withSuperuserClient)
	if err != nil {
		return "", errors.Wrap(err, "not able to filter the super user client.")
	}

	userIndex, err := m.Store.LoadUserIndex()
	if err != nil {
		if err.Error() == "not found" {
			return "No users found in user index", nil
		}
		return "", errors.Wrap(err, "not able to load users from user index.")
	}

	return m.syncUsers(userIndex)
}

func (m *mscalendar) syncUsers(userIndex store.UserIndex) (string, error) {
	if len(userIndex) == 0 {
		return "No connected users found", nil
	}

	numberOfLogs := 0
	users := []*store.User{}
	for _, u := range userIndex {
		// TODO fetch users from kvstore in batches, and process in batches instead of all at once
		user, err := m.Store.LoadUser(u.MattermostUserID)
		if err != nil {
			if numberOfLogs < logTruncateLimit {
				m.Logger.Warnf("Not able to load user %s from user index. err=%v", u.MattermostUserID, err)
			} else if numberOfLogs == logTruncateLimit {
				m.Logger.Warnf(logTruncateMsg)
			}
			numberOfLogs++

			// In case of error in loading, skip this user and continue with the next user
			continue
		}
		if user.Settings.UpdateStatus || user.Settings.ReceiveReminders {
			users = append(users, user)
		}
	}
	if len(users) == 0 {
		return "No users need to be synced", nil
	}

	calendarViews, err := m.GetCalendarViews(users)
	if err != nil {
		return "", errors.Wrap(err, "not able to get calendar views for connected users.")
	}
	if len(calendarViews) == 0 {
		return "No calendar views found", nil
	}

	m.deliverReminders(users, calendarViews)
	out, err := m.setUserStatuses(users, calendarViews)
	if err != nil {
		return "", errors.Wrap(err, "error setting the user statuses.")
	}

	return out, nil
}

func (m *mscalendar) deliverReminders(users []*store.User, calendarViews []*remote.ViewCalendarResponse) {
	numberOfLogs := 0
	toNotify := []*store.User{}
	for _, u := range users {
		if u.Settings.ReceiveReminders {
			toNotify = append(toNotify, u)
		}
	}
	if len(toNotify) == 0 {
		return
	}

	usersByRemoteID := map[string]*store.User{}
	for _, u := range toNotify {
		usersByRemoteID[u.Remote.ID] = u
	}

	for _, view := range calendarViews {
		user, ok := usersByRemoteID[view.RemoteUserID]
		if !ok {
			continue
		}
		if view.Error != nil {
			if numberOfLogs < logTruncateLimit {
				m.Logger.Warnf("Error getting availability for %s. err=%s", user.MattermostUserID, view.Error.Message)
			} else if numberOfLogs == logTruncateLimit {
				m.Logger.Warnf(logTruncateMsg)
			}
			numberOfLogs++
			continue
		}

		mattermostUserID := usersByRemoteID[view.RemoteUserID].MattermostUserID
		m.notifyUpcomingEvents(mattermostUserID, view.Events)
	}
}

func (m *mscalendar) setUserStatuses(users []*store.User, calendarViews []*remote.ViewCalendarResponse) (string, error) {
	numberOfLogs := 0
	toUpdate := []*store.User{}
	for _, u := range users {
		if u.Settings.UpdateStatus {
			toUpdate = append(toUpdate, u)
		}
	}
	if len(toUpdate) == 0 {
		return "No users want their status updated", nil
	}

	mattermostUserIDs := []string{}
	usersByRemoteID := map[string]*store.User{}
	for _, u := range toUpdate {
		mattermostUserIDs = append(mattermostUserIDs, u.MattermostUserID)
		usersByRemoteID[u.Remote.ID] = u
	}

	statuses, appErr := m.PluginAPI.GetMattermostUserStatusesByIds(mattermostUserIDs)
	if appErr != nil {
		return "", errors.Wrap(appErr, "error in getting Mattermost user statuses for connected users.")
	}
	statusMap := map[string]*model.Status{}
	for _, s := range statuses {
		statusMap[s.UserId] = s
	}

	var res string
	for _, view := range calendarViews {
		user, ok := usersByRemoteID[view.RemoteUserID]
		if !ok {
			continue
		}
		if view.Error != nil {
			if numberOfLogs < logTruncateLimit {
				m.Logger.Warnf("Error getting availability for %s. err=%s", user.MattermostUserID, view.Error.Message)
			} else if numberOfLogs == logTruncateLimit {
				m.Logger.Warnf(logTruncateMsg)
			}
			numberOfLogs++
			continue
		}

		mattermostUserID := usersByRemoteID[view.RemoteUserID].MattermostUserID
		status, ok := statusMap[mattermostUserID]
		if !ok {
			continue
		}

		var err error
		res, err = m.setStatusFromCalendarView(user, status, view)
		if err != nil {
			if numberOfLogs < logTruncateLimit {
				m.Logger.Warnf("Error setting user %s status. err=%v", user.MattermostUserID, err)
			} else if numberOfLogs == logTruncateLimit {
				m.Logger.Warnf(logTruncateMsg)
			}
		}
		numberOfLogs++
	}
	if res != "" {
		return res, nil
	}

	return utils.JSONBlock(calendarViews), nil
}

func (m *mscalendar) setStatusFromCalendarView(user *store.User, status *model.Status, res *remote.ViewCalendarResponse) (string, error) {
	currentStatus := status.Status
	if currentStatus == model.STATUS_OFFLINE && !user.Settings.GetConfirmation {
		return "User offline and does not want status change confirmations. No status change", nil
	}

	events := filterBusyEvents(res.Events)
	busyStatus := model.STATUS_DND
	if user.Settings.ReceiveNotificationsDuringMeeting {
		busyStatus = model.STATUS_AWAY
	}

	if len(user.ActiveEvents) == 0 && len(events) == 0 {
		return "No events in local or remote. No status change.", nil
	}

	if len(user.ActiveEvents) > 0 && len(events) == 0 {
		message := fmt.Sprintf("User is no longer busy in calendar, but is not set to busy (%s). No status change.", busyStatus)
		if currentStatus == busyStatus {
			message = "User is no longer busy in calendar. Set status to online."
			if user.LastStatus != "" {
				message = fmt.Sprintf("User is no longer busy in calendar. Set status to previous status (%s)", user.LastStatus)
			}
			err := m.setStatusOrAskUser(user, status, events, true)
			if err != nil {
				return "", errors.Wrapf(err, "error in setting user status for user %s", user.MattermostUserID)
			}
		}

		err := m.Store.StoreUserActiveEvents(user.MattermostUserID, []string{})
		if err != nil {
			return "", errors.Wrapf(err, "error in storing active events for user %s", user.MattermostUserID)
		}
		return message, nil
	}

	remoteHashes := []string{}
	for _, e := range events {
		if e.IsCancelled {
			continue
		}
		h := fmt.Sprintf("%s %s", e.ICalUID, e.Start.Time().UTC().Format(time.RFC3339))
		remoteHashes = append(remoteHashes, h)
	}

	if len(user.ActiveEvents) == 0 {
		var err error
		if currentStatus == busyStatus {
			user.LastStatus = ""
			if status.Manual {
				user.LastStatus = currentStatus
			}
			m.Store.StoreUser(user)
			err = m.Store.StoreUserActiveEvents(user.MattermostUserID, remoteHashes)
			if err != nil {
				return "", errors.Wrapf(err, "error in storing active events for user %s", user.MattermostUserID)
			}
			return "User was already marked as busy. No status change.", nil
		}
		err = m.setStatusOrAskUser(user, status, events, false)
		if err != nil {
			return "", errors.Wrapf(err, "error in setting user status for user %s", user.MattermostUserID)
		}
		err = m.Store.StoreUserActiveEvents(user.MattermostUserID, remoteHashes)
		if err != nil {
			return "", errors.Wrapf(err, "error in storing active events for user %s", user.MattermostUserID)
		}
		return fmt.Sprintf("User was free, but is now busy (%s). Set status to busy.", busyStatus), nil
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

	message := "User is already busy. No status change."
	if currentStatus != busyStatus {
		err := m.setStatusOrAskUser(user, status, events, false)
		if err != nil {
			return "", errors.Wrapf(err, "error in setting user status for user %s", user.MattermostUserID)
		}
		message = fmt.Sprintf("User was free, but is now busy. Set status to busy (%s).", busyStatus)
	}

	err := m.Store.StoreUserActiveEvents(user.MattermostUserID, remoteHashes)
	if err != nil {
		return "", errors.Wrapf(err, "error in storing active events for user %s", user.MattermostUserID)
	}

	return message, nil
}

// setStatusOrAskUser to which status change, and whether it should update the status automatically or ask the user.
// - user: the user to change the status. We use user.LastStatus to determine the status the user had before the beginning of the meeting.
// - currentStatus: currentStatus, to decide whether to store this status when the user is free. This gets assigned to user.LastStatus at the beginning of the meeting.
// - events: the list of events that are triggering this status change
// - isFree: whether the user is free or busy, to decide to which status to change
func (m *mscalendar) setStatusOrAskUser(user *store.User, currentStatus *model.Status, events []*remote.Event, isFree bool) error {
	toSet := model.STATUS_ONLINE
	if isFree && user.LastStatus != "" {
		toSet = user.LastStatus
		user.LastStatus = ""
	}

	if !isFree {
		toSet = model.STATUS_DND
		if user.Settings.ReceiveNotificationsDuringMeeting {
			toSet = model.STATUS_AWAY
		}
		if !user.Settings.GetConfirmation {
			user.LastStatus = ""
			if currentStatus.Manual {
				user.LastStatus = currentStatus.Status
			}
		}
	}

	err := m.Store.StoreUser(user)
	if err != nil {
		return err
	}

	if !user.Settings.GetConfirmation {
		_, appErr := m.PluginAPI.UpdateMattermostUserStatus(user.MattermostUserID, toSet)
		if appErr != nil {
			return appErr
		}
		return nil
	}

	url := fmt.Sprintf("%s%s%s", m.Config.PluginURLPath, config.PathPostAction, config.PathConfirmStatusChange)
	_, err = m.Poster.DMWithAttachments(user.MattermostUserID, views.RenderStatusChangeNotificationView(events, toSet, url))
	if err != nil {
		return err
	}
	return nil
}

func (m *mscalendar) GetCalendarViews(users []*store.User) ([]*remote.ViewCalendarResponse, error) {
	err := m.Filter(withClient)
	if err != nil {
		return nil, err
	}

	start := time.Now().UTC()
	end := time.Now().UTC().Add(calendarViewTimeWindowSize)

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

func (m *mscalendar) notifyUpcomingEvents(mattermostUserID string, events []*remote.Event) {
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
					m.Logger.Warnf("notifyUpcomingEvents error getting timezone. err=%v", err)
					return
				}
			}

			message, err := views.RenderUpcomingEvent(event, timezone)
			if err != nil {
				m.Logger.Warnf("notifyUpcomingEvent error rendering schedule item. err=%v", err)
				continue
			}
			_, err = m.Poster.DM(mattermostUserID, message)
			if err != nil {
				m.Logger.Warnf("notifyUpcomingEvents error creating DM. err=%v", err)
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

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
)

const (
	availabilityTimeWindowSize = 15
)

type Availability interface {
	GetUserAvailabilities(remoteUserID string, scheduleIDs []string) ([]*remote.ScheduleInformation, error)
	SyncStatusForSingleUser(mattermostUserID string) (string, error)
	SyncStatusForAllUsers() (string, error)
}

func (mscalendar *mscalendar) SyncStatusForSingleUser(mattermostUserID string) (string, error) {
	return mscalendar.syncStatusForUsers([]string{mattermostUserID})
}

func (mscalendar *mscalendar) SyncStatusForAllUsers() (string, error) {
	userIndex, err := mscalendar.Store.LoadUserIndex()
	if err != nil {
		if err.Error() == "not found" {
			return "No users found in user index", nil
		}
		return "", err
	}

	mmIDs := userIndex.GetMattermostUserIDs()
	return mscalendar.syncStatusForUsers(mmIDs)
}

func (mscalendar *mscalendar) syncStatusForUsers(mattermostUserIDs []string) (string, error) {
	fullUserIndex, err := mscalendar.Store.LoadUserIndex()
	if err != nil {
		if err.Error() == "not found" {
			return "No users found in user index", nil
		}
		return "", err
	}

	filteredUsers := store.UserIndex{}
	indexByMattermostUserID := fullUserIndex.ByMattermostID()

	for _, mattermostUserID := range mattermostUserIDs {
		if u, ok := indexByMattermostUserID[mattermostUserID]; ok {
			filteredUsers = append(filteredUsers, u)
		}
	}

	if len(filteredUsers) == 0 {
		return "No connected users found", nil
	}

	scheduleIDs := []string{}
	for _, u := range filteredUsers {
		scheduleIDs = append(scheduleIDs, u.Email)
	}

	schedules, err := mscalendar.GetUserAvailabilities(filteredUsers[0].RemoteID, scheduleIDs)
	if err != nil {
		return "", err
	}
	if len(schedules) == 0 {
		return "No schedule info found", nil
	}

	return mscalendar.setUserStatuses(filteredUsers, schedules, mattermostUserIDs)
}

func (mscalendar *mscalendar) setUserStatuses(filteredUsers store.UserIndex, schedules []*remote.ScheduleInformation, mattermostUserIDs []string) (string, error) {
	statuses, appErr := mscalendar.PluginAPI.GetMattermostUserStatusesByIds(mattermostUserIDs)
	if appErr != nil {
		return "", appErr
	}
	statusMap := map[string]string{}
	for _, s := range statuses {
		statusMap[s.UserId] = s.Status
	}

	usersByEmail := filteredUsers.ByEmail()
	var res string
	for _, s := range schedules {
		if s.Error != nil {
			mscalendar.Logger.Errorf("Error getting availability for %s: %s", s.ScheduleID, s.Error.ResponseCode)
			continue
		}

		userID := usersByEmail[s.ScheduleID].MattermostUserID
		status, ok := statusMap[userID]
		if !ok {
			continue
		}

		res = mscalendar.setUserStatusFromAvailability(userID, status, s.AvailabilityView)
	}
	if res != "" {
		return res, nil
	}

	return utils.JSONBlock(schedules), nil
}

func (mscalendar *mscalendar) GetUserAvailabilities(remoteUserID string, scheduleIDs []string) ([]*remote.ScheduleInformation, error) {
	client := mscalendar.MakeSuperuserClient()

	start := remote.NewDateTime(time.Now())
	end := remote.NewDateTime(time.Now().Add(availabilityTimeWindowSize * time.Minute))

	return client.GetSchedule(remoteUserID, scheduleIDs, start, end, availabilityTimeWindowSize)
}

func (mscalendar *mscalendar) setUserStatusFromAvailability(mattermostUserID, currentStatus string, av remote.AvailabilityView) string {
	currentAvailability := av[0]

	switch currentAvailability {
	case remote.AvailabilityViewFree:
		if currentStatus == "dnd" {
			mscalendar.PluginAPI.UpdateMattermostUserStatus(mattermostUserID, "online")
			return fmt.Sprintf("User is free. Setting user from %s to online.", currentStatus)
		} else {
			return fmt.Sprintf("User is free, and is already set to %s.", currentStatus)
		}
	case remote.AvailabilityViewTentative, remote.AvailabilityViewBusy:
		if currentStatus != "dnd" {
			mscalendar.PluginAPI.UpdateMattermostUserStatus(mattermostUserID, "dnd")
			return fmt.Sprintf("User is busy. Setting user from %s to dnd.", currentStatus)
		} else {
			return fmt.Sprintf("User is busy, and is already set to %s.", currentStatus)
		}
	case remote.AvailabilityViewOutOfOffice:
		if currentStatus != "offline" {
			mscalendar.PluginAPI.UpdateMattermostUserStatus(mattermostUserID, "offline")
			return fmt.Sprintf("User is out of office. Setting user from %s to offline", currentStatus)
		} else {
			return fmt.Sprintf("User is out of office, and is already set to %s.", currentStatus)
		}
	case remote.AvailabilityViewWorkingElsewhere:
		return fmt.Sprintf("User is working elsewhere. Pending implementation.")
	}

	return fmt.Sprintf("Availability view doesn't match %d", currentAvailability)
}

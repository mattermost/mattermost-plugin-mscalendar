// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

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

func (api *api) SyncStatusForSingleUser(mattermostUserID string) (string, error) {
	return api.syncStatusForUsers([]string{mattermostUserID})
}

func (api *api) SyncStatusForAllUsers() (string, error) {
	userIndex, err := api.UserStore.LoadUserIndex()
	if err != nil {
		if err.Error() == "not found" {
			return "No users found in user index", nil
		}
		return "", err
	}

	mmIDs := userIndex.GetMattermostUserIDs()
	return api.syncStatusForUsers(mmIDs)
}

func (api *api) syncStatusForUsers(mattermostUserIDs []string) (string, error) {
	fullUserIndex, err := api.UserStore.LoadUserIndex()
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

	schedules, err := api.GetUserAvailabilities(filteredUsers[0].RemoteID, scheduleIDs)
	if err != nil {
		return "", err
	}
	if len(schedules) == 0 {
		return "No schedule info found", nil
	}

	return api.setUserStatuses(filteredUsers, schedules, mattermostUserIDs)
}

func (api *api) setUserStatuses(filteredUsers store.UserIndex, schedules []*remote.ScheduleInformation, mattermostUserIDs []string) (string, error) {
	statuses, appErr := api.Dependencies.PluginAPI.GetUserStatusesByIds(mattermostUserIDs)
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
			api.Logger.Errorf("Error getting availability for %s: %s", s.ScheduleID, s.Error.ResponseCode)
			continue
		}

		userID := usersByEmail[s.ScheduleID].MattermostUserID
		status, ok := statusMap[userID]
		if !ok {
			continue
		}

		res = api.setUserStatusFromAvailability(userID, status, s.AvailabilityView)
	}
	if res != "" {
		return res, nil
	}

	return utils.JSONBlock(schedules), nil
}

func (api *api) GetUserAvailabilities(remoteUserID string, scheduleIDs []string) ([]*remote.ScheduleInformation, error) {
	client := api.MakeSuperuserClient()

	start := remote.NewDateTime(time.Now().UTC(), "UTC")
	end := remote.NewDateTime(time.Now().UTC().Add(availabilityTimeWindowSize*time.Minute), "UTC")

	return client.GetSchedule(remoteUserID, scheduleIDs, start, end, availabilityTimeWindowSize)
}

func (api *api) setUserStatusFromAvailability(mattermostUserID, currentStatus string, av remote.AvailabilityView) string {
	currentAvailability := av[0]

	switch currentAvailability {
	case remote.AvailabilityViewFree:
		if currentStatus == "dnd" {
			api.PluginAPI.UpdateUserStatus(mattermostUserID, "online")
			return fmt.Sprintf("User is free. Setting user from %s to online.", currentStatus)
		} else {
			return fmt.Sprintf("User is free, and is already set to %s.", currentStatus)
		}
	case remote.AvailabilityViewTentative, remote.AvailabilityViewBusy:
		if currentStatus != "dnd" {
			api.PluginAPI.UpdateUserStatus(mattermostUserID, "dnd")
			return fmt.Sprintf("User is busy. Setting user from %s to dnd.", currentStatus)
		} else {
			return fmt.Sprintf("User is busy, and is already set to %s.", currentStatus)
		}
	case remote.AvailabilityViewOutOfOffice:
		if currentStatus != "offline" {
			api.PluginAPI.UpdateUserStatus(mattermostUserID, "offline")
			return fmt.Sprintf("User is out of office. Setting user from %s to offline", currentStatus)
		} else {
			return fmt.Sprintf("User is out of office, and is already set to %s.", currentStatus)
		}
	case remote.AvailabilityViewWorkingElsewhere:
		return fmt.Sprintf("User is working elsewhere. Pending implementation.")
	}

	return fmt.Sprintf("Availability view doesn't match %d", currentAvailability)
}

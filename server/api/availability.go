// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
	"github.com/pkg/errors"
)

const (
	availabilityTimeWindowSize = 15
)

func (api *api) SyncStatusForSingleUser(mattermostUserID string) (string, error) {
	u, err := api.UserStore.LoadUser(mattermostUserID)
	if err != nil {
		return "", err
	}

	scheduleIDs := []string{u.Remote.Mail}
	sched, err := api.GetUserAvailabilities(u.Remote.ID, scheduleIDs)

	if err != nil {
		return "", err
	}
	if len(sched) == 0 {
		return "No schedule info found", nil
	}

	status, appErr := api.Dependencies.API.GetUserStatus(api.mattermostUserID)
	if appErr != nil {
		return "", appErr
	}

	s := sched[0]
	if s.Error != nil {
		return "", errors.Errorf("Error getting availability for %s: %s", s.ScheduleID, s.Error.ResponseCode)
	}
	if len(s.AvailabilityView) == 0 {
		return "No availabilities found", nil
	}

	av := s.AvailabilityView[0]
	return api.setUserStatusFromAvailability(api.mattermostUserID, status.Status, av), nil
}

func (api *api) SyncStatusForAllUsers() (string, error) {
	userIndex, err := api.UserStore.LoadUserIndex()
	if err != nil {
		if err.Error() == "not found" {
			return "No users found in user index", nil
		}
		return "", err
	}

	if len(userIndex) == 0 {
		return "No connected users found", nil
	}

	scheduleIDs := []string{}
	mattermostUserIDs := []string{}
	for _, u := range userIndex {
		scheduleIDs = append(scheduleIDs, u.Email)
		mattermostUserIDs = append(mattermostUserIDs, u.MattermostUserID)
	}

	sched, err := api.GetUserAvailabilities(userIndex[0].RemoteID, scheduleIDs)
	if err != nil {
		return "", err
	}
	if len(sched) == 0 {
		return "No schedule info found", nil
	}

	statuses, appErr := api.Dependencies.API.GetUserStatusesByIds(mattermostUserIDs)
	if appErr != nil {
		return "", appErr
	}

	statusMap := map[string]string{}
	for _, s := range statuses {
		statusMap[s.UserId] = s.Status
	}

	usersByEmail := userIndex.ByEmail()

	var res string
	for _, s := range sched {
		if s.Error != nil {
			api.Logger.Errorf("Error getting availability for %s: %s", s.ScheduleID, s.Error.ResponseCode)
			continue
		}

		av := s.AvailabilityView[0]

		userID := usersByEmail[s.ScheduleID].MattermostUserID
		status, ok := statusMap[userID]
		if !ok {
			continue
		}

		res = api.setUserStatusFromAvailability(userID, status, av)
	}

	if res != "" {
		return res, nil
	}

	return utils.JSONBlock(sched), nil
}

func (api *api) GetUserAvailabilities(remoteUserID string, scheduleIDs []string) ([]*remote.ScheduleInformation, error) {
	client := api.MakeSuperuserClient()

	start := remote.NewDateTime(time.Now())
	end := remote.NewDateTime(time.Now().Add(availabilityTimeWindowSize * time.Minute))

	return client.GetSchedule(remoteUserID, scheduleIDs, start, end, availabilityTimeWindowSize)
}

func (api *api) setUserStatusFromAvailability(mattermostUserID, currentStatus string, av byte) string {
	switch av {
	case remote.AvailabilityViewFree:
		if currentStatus == "dnd" {
			api.API.UpdateUserStatus(mattermostUserID, "online")
			return fmt.Sprintf("User is free. Setting user from %s to online.", currentStatus)
		} else {
			return fmt.Sprintf("User is free, and is already set to %s.", currentStatus)
		}
	case remote.AvailabilityViewTentative, remote.AvailabilityViewBusy:
		if currentStatus != "dnd" {
			api.API.UpdateUserStatus(mattermostUserID, "dnd")
			return fmt.Sprintf("User is busy. Setting user from %s to dnd.", currentStatus)
		} else {
			return fmt.Sprintf("User is busy, and is already set to %s.", currentStatus)
		}
	case remote.AvailabilityViewOutOfOffice:
		if currentStatus != "offline" {
			api.API.UpdateUserStatus(mattermostUserID, "offline")
			return fmt.Sprintf("User is out of office. Setting user from %s to offline", currentStatus)
		} else {
			return fmt.Sprintf("User is out of office, and is already set to %s.", currentStatus)
		}
	case remote.AvailabilityViewWorkingElsewhere:
		return fmt.Sprintf("User is working elsewhere. Pending implementation.")
	}

	return fmt.Sprintf("Availability view doesn't match %d", av)
}

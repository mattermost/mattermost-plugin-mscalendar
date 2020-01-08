// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
)

const (
	availabilityTimeWindowSize = 15

	availabilityViewFree             = '0'
	availabilityViewTentative        = '1'
	availabilityViewBusy             = '2'
	availabilityViewOutOfOffice      = '3'
	availabilityViewWorkingElsewhere = '4'
)

func (api *api) SyncStatusForSingleUser() (string, error) {
	u, err := api.UserStore.LoadUser(api.mattermostUserID)
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

	s := sched[0]
	av := s.AvailabilityView[0]
	return api.setUserStatusFromAvailability(api.mattermostUserID, av), nil
}

func (api *api) SyncStatusForAllUsers() (string, error) {
	users, err := api.UserStore.LoadAllUsers()
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

	sched, err := api.GetUserAvailabilities(users[0].RemoteID, scheduleIDs)
	if err != nil {
		return "", err
	}
	if len(sched) == 0 {
		return "No schedule info found", nil
	}

	var res string
	for i, s := range sched {
		userID := users[i].MattermostUserID
		av := s.AvailabilityView[0]
		res = api.setUserStatusFromAvailability(userID, av)
	}

	if res != "" {
		return res, nil
	}

	return utils.JSONBlock(sched), nil
}

func (api *api) GetUserAvailabilities(remoteUserID string, scheduleIDs []string) ([]*remote.ScheduleInformation, error) {
	client, err := api.MakeSuperuserClient()
	if err != nil {
		return nil, err
	}

	start := remote.NewDateTime(time.Now())
	end := remote.NewDateTime(time.Now().Add(availabilityTimeWindowSize * time.Minute))

	return client.GetSchedule(remoteUserID, scheduleIDs, start, end, availabilityTimeWindowSize)
}

func (api *api) setUserStatusFromAvailability(mattermostUserID string, av byte) string {
	currentStatus, _ := api.API.GetUserStatus(mattermostUserID)

	switch av {
	case availabilityViewFree:
		if currentStatus.Status == "dnd" {
			api.API.UpdateUserStatus(mattermostUserID, "online")
			return fmt.Sprintf("User is free. Setting user from %s to online.", currentStatus.Status)
		} else {
			return fmt.Sprintf("User is free, and is already set to %s.", currentStatus.Status)
		}
	case availabilityViewTentative, availabilityViewBusy:
		if currentStatus.Status != "dnd" {
			api.API.UpdateUserStatus(mattermostUserID, "dnd")
			return fmt.Sprintf("User is busy. Setting user from %s to dnd.", currentStatus.Status)
		} else {
			return fmt.Sprintf("User is busy, and is already set to %s.", currentStatus.Status)
		}
	case availabilityViewOutOfOffice:
		if currentStatus.Status != "offline" {
			api.API.UpdateUserStatus(mattermostUserID, "offline")
			return fmt.Sprintf("User is out of office. Setting user from %s to offline", currentStatus.Status)
		} else {
			return fmt.Sprintf("User is out of office, and is already set to %s.", currentStatus.Status)
		}
	case availabilityViewWorkingElsewhere:
		return fmt.Sprintf("User is working elsewhere. Pending implementation.")
	}

	return fmt.Sprintf("Availability view doesn't match %d", av)
}

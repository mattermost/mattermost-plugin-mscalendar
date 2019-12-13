// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
)

const (
	AVAILABILITY_VIEW_FREE              = '0'
	AVAILABILITY_VIEW_TENTATIVE         = '1'
	AVAILABILITY_VIEW_BUSY              = '2'
	AVAILABILITY_VIEW_OUT_OF_OFFICE     = '3'
	AVAILABILITY_VIEW_WORKING_ELSEWHERE = '4'
)

func (api *api) GetUserAvailability() (string, error) {
	client, err := api.MakeClient()
	if err != nil {
		return "", err
	}

	users, err := api.UserStore.LoadAllUsers()
	if err != nil {
		return "", err
	}

	scheduleIDs := []string{}
	for _, u := range users {
		scheduleIDs = append(scheduleIDs, u.Email)
	}

	start := remote.NewDateTime(time.Now())
	end := remote.NewDateTime(time.Now().Add(15 * time.Minute))
	timeWindow := 15 // minutes
	sched, err := client.GetSchedule(scheduleIDs, start, end, timeWindow)
	if err != nil {
		return "", err
	}

	userID := users[0].MattermostUserID
	av := sched[0].AvailabilityView

	setUserStatusFromAvailability(api, userID, av[0])

	return utils.JSONBlock(sched), err
}

func setUserStatusFromAvailability(api *api, mattermostUserID string, av byte) {
	currentStatus, _ := api.API.GetUserStatus(mattermostUserID)

	switch av {
	case AVAILABILITY_VIEW_FREE:
		if currentStatus.Status == "dnd" {
			api.Logger.Debugf("Setting user to online")
			api.API.UpdateUserStatus(mattermostUserID, "online")
		} else {
			api.Logger.Debugf("User is already online")
		}
	case AVAILABILITY_VIEW_TENTATIVE, AVAILABILITY_VIEW_BUSY:
		if currentStatus.Status != "dnd" {
			api.Logger.Debugf("Setting user to dnd")
			api.API.UpdateUserStatus(mattermostUserID, "dnd")
		} else {
			api.Logger.Debugf("User is already dnd")
		}
	case AVAILABILITY_VIEW_OUT_OF_OFFICE:
		if currentStatus.Status != "offline" {
			api.Logger.Debugf("Setting user to out of office")
			api.API.UpdateUserStatus(mattermostUserID, "offline")
		} else {
			api.Logger.Debugf("User is already offline")
		}
	case AVAILABILITY_VIEW_WORKING_ELSEWHERE:
		if currentStatus.Status != "dnd" {
			api.Logger.Debugf("Setting user to working elsewhere")
			api.API.UpdateUserStatus(mattermostUserID, "online")
		} else {
			api.Logger.Debugf("User is already online")
		}
	default:
		api.Logger.Debugf("Availability view doesn't match", "av", av)
	}
}

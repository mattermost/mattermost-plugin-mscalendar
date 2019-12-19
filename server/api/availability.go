// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/mattermost/mattermost-plugin-msoffice/server/job"
	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils"
	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/bot"
)

const (
	AVAILABILITY_VIEW_FREE              = '0'
	AVAILABILITY_VIEW_TENTATIVE         = '1'
	AVAILABILITY_VIEW_BUSY              = '2'
	AVAILABILITY_VIEW_OUT_OF_OFFICE     = '3'
	AVAILABILITY_VIEW_WORKING_ELSEWHERE = '4'
)

type availabilityJob struct {
	api API
}

func NewAvailabilityJob(api API) job.RecurringJob {
	return &availabilityJob{api: api}
}

func (j *availabilityJob) Run() {
	c := cron.New()
	c.AddFunc("* * * * *", j.Work)
	c.Start()
}

func (j *availabilityJob) getLogger() bot.Logger {
	return j.api.(*api).Logger
}

func (j *availabilityJob) Work() {
	log := j.getLogger()
	log.Debugf("Availability job beginning")

	_, err := j.api.GetUserAvailability()
	if err != nil {
		log.Errorf("Error during Availability job", "error", err.Error())
	}

	log.Debugf("Availability job finished")
}

func (api *api) GetUserAvailability() (string, error) {
	client, err := api.MakeClient()
	if err != nil {
		return "", err
	}

	u, err := api.UserStore.LoadUser(api.mattermostUserID)
	if err != nil {
		return "", err
	}

	scheduleIDs := []string{u.Remote.Mail}

	start, end, timeWindow := getTimeInfoForAvailability()

	sched, err := client.GetSchedule(u.Remote.ID, scheduleIDs, start, end, timeWindow)
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

func (api *api) GetAllUsersAvailability() (string, error) {
	client, err := api.MakeAppClient()
	if err != nil {
		return "", err
	}

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

	start, end, timeWindow := getTimeInfoForAvailability()

	sched, err := client.GetSchedule(users[0].RemoteID, scheduleIDs, start, end, timeWindow)
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

func getTimeInfoForAvailability() (start, end *remote.DateTime, timeWindow int) {
	start = remote.NewDateTime(time.Now())
	end = remote.NewDateTime(time.Now().Add(15 * time.Minute))
	timeWindow = 15 // minutes
	return start, end, timeWindow
}

func (api *api) setUserStatusFromAvailability(mattermostUserID string, av byte) string {
	currentStatus, _ := api.API.GetUserStatus(mattermostUserID)

	switch av {
	case AVAILABILITY_VIEW_FREE:
		if currentStatus.Status == "dnd" {
			api.API.UpdateUserStatus(mattermostUserID, "online")
			return fmt.Sprintf("User is free. Setting user from %s to online.", currentStatus.Status)
		} else {
			return fmt.Sprintf("User is free, and is already set to %s.", currentStatus.Status)
		}
	case AVAILABILITY_VIEW_TENTATIVE, AVAILABILITY_VIEW_BUSY:
		if currentStatus.Status != "dnd" {
			api.API.UpdateUserStatus(mattermostUserID, "dnd")
			return fmt.Sprintf("User is busy. Setting user from %s to dnd.", currentStatus.Status)
		} else {
			return fmt.Sprintf("User is busy, and is already set to %s.", currentStatus.Status)
		}
	case AVAILABILITY_VIEW_OUT_OF_OFFICE:
		if currentStatus.Status != "offline" {
			api.API.UpdateUserStatus(mattermostUserID, "offline")
			return fmt.Sprintf("User is out of office. Setting user from %s to offline", currentStatus.Status)
		} else {
			return fmt.Sprintf("User is out of office, and is already set to %s.", currentStatus.Status)
		}
	case AVAILABILITY_VIEW_WORKING_ELSEWHERE:
		return fmt.Sprintf("User is working elsewhere. Pending implementation.")
	}

	return fmt.Sprintf("Availability view doesn't match %d", av)
}

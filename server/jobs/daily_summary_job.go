// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package jobs

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
)

// Unique id for the daily summary job
const dailySummaryJobID = "daily_summary"

const dailySummaryJobInterval = mscalendar.DailySummaryJobInterval

// NewDailySummaryJob creates a RegisteredJob with the parameters specific to the DailySummaryJob
func NewDailySummaryJob() RegisteredJob {
	return RegisteredJob{
		id:                dailySummaryJobID,
		interval:          dailySummaryJobInterval,
		work:              runDailySummaryJob,
		isEnabledByConfig: isDailySummaryJobEnabled,
	}
}

// runDailySummaryJob delivers the daily calendar summary to all users who have their settings configured to receive it now
func runDailySummaryJob(env mscalendar.Env) {
	mscal := mscalendar.New(env, "")
	_, err := mscal.GetRemoteUser(env.BotUserID)
	if err != nil {
		env.Logger.Errorf("Please connect bot user using `/%s connect_bot`, in order to run the daily summary job.", config.CommandTrigger)
		return
	}

	env.Logger.Debugf("Daily summary job beginning")

	err = mscal.ProcessAllDailySummary()
	if err != nil {
		env.Logger.Errorf("Error during daily summary job", "error", err.Error())
	}

	env.Logger.Debugf("Daily summary job finished")
}

// isDailySummaryJobEnabled uses current config to determine whether the job is enabled.
func isDailySummaryJobEnabled(env mscalendar.Env) bool {
	return env.EnableDailySummary
}

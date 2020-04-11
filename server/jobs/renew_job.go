// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package jobs

import (
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
)

const ditherRenew = 50 * time.Millisecond

func NewRenewJob() RegisteredJob {
	return RegisteredJob{
		id:       "renew",
		interval: 24 * time.Hour,
		work:     runRenewJob,
	}
}

// runDailySummaryJob delivers the daily calendar summary to all users who have their settings configured to receive it now
func runRenewJob(env mscalendar.Env) {
	uindex, err := env.Store.LoadUserIndex()
	if err != nil {
		env.Logger.Errorf("Renew job failed to load user index: %s", err.Error())
		return
	}
	env.Logger.Debugf("Renew job: %v users", len(uindex))

	for _, u := range uindex {
		asUser := mscalendar.New(env, u.MattermostUserID)
		env.Logger.Debugf("Renewing for user: %s", u.MattermostUserID)
		_, err = asUser.RenewMyEventSubscription()
		if err != nil {
			env.Logger.Errorf("Error renewing subscription: %s", err.Error())
		}

		time.Sleep(ditherRenew)
	}

	env.Logger.Debugf("Renew job finished")
}

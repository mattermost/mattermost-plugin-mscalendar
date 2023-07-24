// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package jobs

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
)

// Unique id for the status sync job
const statusSyncJobID = "status_sync"

// NewStatusSyncJob creates a RegisteredJob with the parameters specific to the StatusSyncJob
func NewStatusSyncJob() RegisteredJob {
	return RegisteredJob{
		id:       statusSyncJobID,
		interval: mscalendar.StatusSyncJobInterval,
		work:     runSyncJob,
	}
}

// runSyncJob synchronizes all users' statuses between mscalendar and Mattermost.
func runSyncJob(env mscalendar.Env) {
	env.Logger.Debugf("User status sync job beginning")

	_, syncJobSummary, err := mscalendar.New(env, "").SyncAll()
	if err != nil {
		env.Logger.Errorf("Error during user status sync job. err=%v", err)
	}

	env.Logger.Debugf("User status sync job finished.\nSummary\nNumber of users processed:- %d\nNumber of users had their status changed:- %d\nNumber of users had errors:- %d", syncJobSummary.NumberOfUsersProcessed, syncJobSummary.NumberOfUsersStatusChanged, syncJobSummary.NumberOfUsersFailedStatusChanged)
}

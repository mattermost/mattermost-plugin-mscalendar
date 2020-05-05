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
		id:                statusSyncJobID,
		interval:          mscalendar.StatusSyncJobInterval,
		work:              runSyncJob,
		isEnabledByConfig: isStatusSyncJobEnabled,
	}
}

// runSyncJob synchronizes all users' statuses between mscalendar and Mattermost.
func runSyncJob(env mscalendar.Env) {
	env.Logger.Debugf("User status sync job beginning")

	_, err := mscalendar.New(env, "").SyncAll()
	if err != nil {
		env.Logger.Errorf("Error during user status sync job: %v", err)
	}

	env.Logger.Debugf("User status sync job finished")
}

// isStatusSyncJobEnabled uses current config to determine whether the job is enabled.
func isStatusSyncJobEnabled(env mscalendar.Env) bool {
	return env.EnableStatusSync
}

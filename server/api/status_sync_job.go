// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"sync"
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/utils/bot"
)

type StatusSyncJob struct {
	api        API
	cancel     chan struct{}
	cancelled  chan struct{}
	cancelOnce sync.Once
}

func (j *StatusSyncJob) getLogger() bot.Logger {
	return j.api.(*api).Logger
}

func (j *StatusSyncJob) work() {
	log := j.getLogger()
	log.Debugf("User status sync job beginning")

	_, err := j.api.SyncStatusForAllUsers()
	if err != nil {
		log.Errorf("Error during user status sync job", "error", err.Error())
	}

	log.Debugf("User status sync job finished")
}

func NewStatusSyncJob(api API) *StatusSyncJob {
	return &StatusSyncJob{
		cancel:    make(chan struct{}),
		cancelled: make(chan struct{}),
		api:       api,
	}
}

const JOB_INTERVAL = 1 * time.Minute

func (job *StatusSyncJob) Start() {
	go func() {
		defer close(job.cancelled)

		ticker := time.NewTicker(JOB_INTERVAL)
		defer func() {
			ticker.Stop()
		}()

		for {
			select {
			case <-ticker.C:
				job.work()
			case <-job.cancel:
				return
			}
		}
	}()
}

func (job *StatusSyncJob) Cancel() {
	job.cancelOnce.Do(func() {
		close(job.cancel)
	})
	<-job.cancelled
}

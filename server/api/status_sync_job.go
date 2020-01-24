// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"sync"
	"time"
)

const JOB_INTERVAL = 5 * time.Minute

type StatusSyncJob struct {
	api        API
	cancel     chan struct{}
	cancelled  chan struct{}
	cancelOnce sync.Once
}

func (j *StatusSyncJob) work() {
	api := j.api
	api.Debugf("User status sync job beginning")

	_, err := j.api.SyncStatusForAllUsers()
	if err != nil {
		api.Errorf("Error during user status sync job", "error", err.Error())
	}

	api.Debugf("User status sync job finished")
}

func NewStatusSyncJob(api API) *StatusSyncJob {
	return &StatusSyncJob{
		cancel:    make(chan struct{}),
		cancelled: make(chan struct{}),
		api:       api,
	}
}

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

// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"sync"
	"time"
)

const JOB_INTERVAL = 5 * time.Minute

type StatusSyncJob struct {
	Env

	cancel     chan struct{}
	cancelled  chan struct{}
	cancelOnce sync.Once
}

func NewStatusSyncJob(env Env) *StatusSyncJob {
	return &StatusSyncJob{
		cancel:    make(chan struct{}),
		cancelled: make(chan struct{}),
		Env:       env,
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

func (j *StatusSyncJob) work() {
	j.Logger.Debugf("User status sync job beginning")

	_, err := New(j.Env, "").SyncStatusAll()
	if err != nil {
		j.Logger.Errorf("Error during user status sync job", "error", err.Error())
	}

	j.Logger.Debugf("User status sync job finished")
}

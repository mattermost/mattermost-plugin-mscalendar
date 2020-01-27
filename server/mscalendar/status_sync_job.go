// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"sync"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
)

const JOB_INTERVAL = 5 * time.Minute

type StatusSyncJob struct {
	mscalendar MSCalendar
	cancel     chan struct{}
	cancelled  chan struct{}
	cancelOnce sync.Once
}

func (j *StatusSyncJob) getLogger() bot.Logger {
	return j.mscalendar.(*mscalendar).Logger
}

func (j *StatusSyncJob) work() {
	log := j.getLogger()
	log.Debugf("User status sync job beginning")

	_, err := j.mscalendar.SyncStatusForAllUsers()
	if err != nil {
		log.Errorf("Error during user status sync job", "error", err.Error())
	}

	log.Debugf("User status sync job finished")
}

func NewStatusSyncJob(mscalendar MSCalendar) *StatusSyncJob {
	return &StatusSyncJob{
		cancel:     make(chan struct{}),
		cancelled:  make(chan struct{}),
		mscalendar: mscalendar,
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
